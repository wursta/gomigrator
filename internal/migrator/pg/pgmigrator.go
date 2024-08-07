package pgmigrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

	// Пропускаем, если не удалось создать таблицу по причине ошибки
	// "ERROR: duplicate key value violates unique constraint... (SQLSTATE 23505)"
	// Это значит, что таблица уже создана.
	var e *pgconn.PgError
	if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
		return nil
	}
	if err != nil {
		return err
	}

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

		migrated, err := m.ApplyMigration(
			ctx,
			migrations[i],
			migrator.MigrationDirectionUp,
			migrator.MigrationStatusMigrating,
			migrator.MigrationStatusMigrated,
			migrator.MigrationStatusFailed,
			func(ctx context.Context, db *sqlx.DB, migration migrator.Migration) bool {
				// Если взяли lock, надо ещё раз проверить, что миграция ещё не выполнена.
				return !isMigrated(ctx, db, migration.Name)
			},
		)
		if err != nil {
			return fmt.Errorf("error while up migration %s: %w", migrations[i].Name, err)
		}
		if migrated {
			log.Println("Success migration:", migrations[i].Name)
		} else {
			log.Println("Skip migration:", migrations[i].Name)
		}
	}
	return nil
}

func (m *PgMigrator) Down(ctx context.Context, migrations []migrator.Migration) error {
	for i := range migrations {
		log.Println("Start rollback:", migrations[i].Name)

		_, err := m.ApplyMigration(
			ctx,
			migrations[i],
			migrator.MigrationDirectionDown,
			migrator.MigrationStatusMigrating,
			migrator.MigrationStatusNew,
			migrator.MigrationStatusMigrated,
			nil,
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
	afterLockCheck func(ctx context.Context, db *sqlx.DB, migration migrator.Migration) bool,
) (bool, error) {
	err := lock(ctx, m.db, migration)
	if err != nil {
		return false, fmt.Errorf("error while locking migration row: %w", err)
	}
	defer unlock(ctx, m.db, migration.Name)

	if afterLockCheck != nil {
		if !afterLockCheck(ctx, m.db, migration) {
			return false, nil
		}
	}

	err = updateMigrationStatus(ctx, m.db, migration, startStatus)
	if err != nil {
		return false, err
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
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
			return false, updateStatusErr
		}
		return false, err
	}

	tx.Commit()

	err = updateMigrationStatus(ctx, m.db, migration, successStatus)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *PgMigrator) GetLastMigations(
	ctx context.Context,
	status migrator.MigrationStatus,
	count int,
) ([]migrator.Migration, error) {
	rows, err := m.db.NamedQueryContext(
		ctx,
		`SELECT id, file_path, file_name, status, create_dt, migrate_dt 
		FROM public.dbmigrations 
		WHERE status = :migration_status
		ORDER BY id DESC 
		LIMIT :count`,
		map[string]interface{}{
			"migration_status": status,
			"count":            count,
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

func lock(ctx context.Context, db *sqlx.DB, migration migrator.Migration) error {
	_, err := db.NamedExecContext(
		ctx,
		`INSERT INTO public.dbmigrations (file_path, file_name, status, create_dt, migrate_dt) 
		 VALUES (:migration_filepath, :migration_filename, :migration_status, NOW(), NOW())`,
		map[string]interface{}{
			"migration_filepath": migration.FilePath,
			"migration_filename": migration.Name,
			"migration_status":   migrator.MigrationStatusNew,
		},
	)

	// Пропускаем ошибку, есл это "ERROR: duplicate key value violates unique constraint... (SQLSTATE 23505)".
	// Значит запись в таблице уже есть.
	var e *pgconn.PgError
	if errors.As(err, &e) && e.Code != pgerrcode.UniqueViolation {
		return err
	}

	_, err = db.NamedExecContext(
		ctx,
		`SELECT pg_advisory_lock(id) FROM public.dbmigrations WHERE file_name = :migration_filename`,
		map[string]interface{}{
			"migration_filename": migration.Name,
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
		`UPDATE public.dbmigrations 
		SET
			status = :migration_status,
		 	migrate_dt = NOW()
		WHERE file_name = :migration_filename`,
		map[string]interface{}{
			"migration_filename": migration.Name,
			"migration_status":   migrationStatus,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
