package pgmigrator

import (
	"context"
	"fmt"
	"log"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/wursta/gomigrator/internal/migrator"
)

type PgMigrator struct {
	dsn              string
	migrationsSchema string
	migratinsTable   string
	mu               sync.Mutex
	db               *sqlx.DB
}

func New(dsn string) *PgMigrator {
	return &PgMigrator{
		dsn:              dsn,
		migrationsSchema: "public",
		migratinsTable:   "dbmigrations",
	}
}

func (m *PgMigrator) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db != nil {
		// already connected
		return nil
	}

	db, err := sqlx.ConnectContext(ctx, "pgx", m.dsn)
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	m.db = db

	return nil
}

func (m *PgMigrator) Close() error {
	if m.db == nil {
		return nil
	}

	err := m.db.Close()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.db = nil
	return nil
}

func (m *PgMigrator) Init(ctx context.Context) error {
	row := m.db.QueryRowxContext(
		ctx,
		`SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE  table_schema = 'public'
			AND    table_name   = 'dbmigrations'
		) as exists`,
	)

	var migrationsTableExists bool
	err := row.Scan(&migrationsTableExists)
	if err != nil {
		return fmt.Errorf("error while scan existence row: %w", err)
	}

	if migrationsTableExists {
		return nil
	}

	_, err = m.db.ExecContext(
		ctx,
		`CREATE TABLE public.dbmigrations (
			id SERIAL NOT NULL,
			file_path VARCHAR(255) NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			status VARCHAR(10) NOT NULL,
			create_dt timestamp NOT NULL,
			migrate_dt timestamp NOT NULL,
			CONSTRAINT dbmigrations_pk PRIMARY KEY (id)
		)`,
	)
	if err != nil {
		return fmt.Errorf("error while create migrations log table: %w", err)
	}

	_, err = m.db.ExecContext(
		ctx,
		`CREATE UNIQUE INDEX unq_dbmigrations_file_name ON public.dbmigrations (file_name)`,
	)
	if err != nil {
		return fmt.Errorf("error while create index on migrations log table: %w", err)
	}

	return nil
}

func (m *PgMigrator) Up(ctx context.Context, migrations []migrator.Migration) error {
	for i := range migrations {
		log.Println("Start migration:", migrations[i].Name)

		if isMigrated(ctx, m.db, migrations[i].Name) {
			log.Println("Skip migration:", migrations[i].Name)
			continue
		}

		err := m.ApplyMigration(
			ctx,
			migrations[i],
			migrator.MigrationDirectionUp,
			migrator.MigrationStatusMigrating,
			migrator.MigrationStatusMigrated,
			migrator.MigrationStatusFailed,
		)
		if err != nil {
			return fmt.Errorf("error while up migration %s: %w", migrations[i].Name, err)
		}
		log.Println("Success migration:", migrations[i].Name)
	}
	return nil
}

func (m *PgMigrator) Down(ctx context.Context, migrations []migrator.Migration) error {
	for i := range migrations {
		log.Println("Start rollback:", migrations[i].Name)

		err := m.ApplyMigration(
			ctx,
			migrations[i],
			migrator.MigrationDirectionDown,
			migrator.MigrationStatusMigrating,
			migrator.MigrationStatusNew,
			migrator.MigrationStatusMigrated,
		)
		if err != nil {
			return fmt.Errorf("error while rollback migration %s: %w", migrations[i].Name, err)
		}
		log.Println("Success rollback:", migrations[i].Name)
	}
	return nil
}

func (m *PgMigrator) ApplyMigration(
	ctx context.Context,
	migration migrator.Migration,
	direction migrator.MigrationDirection,
	startStatus,
	successStatus,
	failStatus migrator.MigrationStatus,
) error {
	err := lock(ctx, m.db, migration.Name)
	if err != nil {
		return fmt.Errorf("error while locking migration row: %w", err)
	}
	defer unlock(ctx, m.db, migration.Name)

	err = updateMigrationStatus(ctx, m.db, migration, startStatus)
	if err != nil {
		return err
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if direction == migrator.MigrationDirectionUp {
		err = migration.UpHandlerContext(ctx, tx)
	} else {
		err = migration.DownHandlerContext(ctx, tx)
	}
	if err != nil {
		tx.Rollback()
		updateStatusErr := updateMigrationStatus(ctx, m.db, migration, failStatus)
		if updateStatusErr != nil {
			return updateStatusErr
		}
		return err
	}

	tx.Commit()

	err = updateMigrationStatus(ctx, m.db, migration, successStatus)
	if err != nil {
		return err
	}

	return nil
}

func (m *PgMigrator) GetLastMigations(ctx context.Context, count int) ([]migrator.Migration, error) {
	rows, err := m.db.NamedQueryContext(
		ctx,
		`SELECT id, file_path, file_name, status, create_dt, migrate_dt 
		FROM public.dbmigrations 
		ORDER BY id DESC 
		LIMIT :count`,
		map[string]interface{}{
			"count": count,
		},
	)
	if err != nil {
		return nil, err
	}

	migrations := []migrator.Migration{}
	for rows.Next() {
		var migration migrator.Migration
		err = rows.StructScan(&migration)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)
	}
	return migrations, nil
}

func (m *PgMigrator) GetMigations(ctx context.Context) ([]migrator.Migration, error) {
	rows, err := m.db.NamedQueryContext(
		ctx,
		`SELECT id, file_path, file_name, status, create_dt, migrate_dt
		FROM public.dbmigrations 
		ORDER BY id DESC`,
		map[string]interface{}{},
	)
	if err != nil {
		return nil, err
	}

	migrations := []migrator.Migration{}
	for rows.Next() {
		var migration migrator.Migration
		err = rows.StructScan(&migration)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)
	}
	return migrations, nil
}

func isMigrated(ctx context.Context, db *sqlx.DB, migrationFileName string) bool {
	rows, err := db.NamedQueryContext(
		ctx,
		"SELECT status FROM public.dbmigrations WHERE file_name = :migration_filename",
		map[string]interface{}{
			"migration_filename": migrationFileName,
		},
	)
	if err != nil {
		return false
	}

	for rows.Next() {
		var status migrator.MigrationStatus
		rows.Scan(&status)
		return status == migrator.MigrationStatusMigrated || status == migrator.MigrationStatusMigrating
	}
	return false
}

func lock(ctx context.Context, db *sqlx.DB, migrationName string) error {
	_, err := db.NamedExecContext(
		ctx,
		`SELECT pg_advisory_lock(id) FROM public.dbmigrations WHERE file_name = :migration_filename`,
		map[string]interface{}{
			"migration_filename": migrationName,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func unlock(ctx context.Context, db *sqlx.DB, migrationName string) {
	db.NamedExecContext(
		ctx,
		`SELECT pg_advisory_unlock(id) FROM public.dbmigrations WHERE file_name = :migration_filename`,
		map[string]interface{}{
			"migration_filename": migrationName,
		},
	)
}

func updateMigrationStatus(
	ctx context.Context,
	db *sqlx.DB,
	migration migrator.Migration,
	migrationStatus migrator.MigrationStatus,
) error {
	_, err := db.NamedExecContext(
		ctx,
		`INSERT INTO public.dbmigrations (file_path, file_name, status, create_dt, migrate_dt) 
		 VALUES (:migration_filepath, :migration_filename, :migration_status, NOW(), NOW())
		 ON CONFLICT (file_name) DO UPDATE SET
		 status = :migration_status,
		 migrate_dt = NOW()`,
		map[string]interface{}{
			"migration_filepath": migration.FilePath,
			"migration_filename": migration.Name,
			"migration_status":   migrationStatus,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
