package pgmigrator

import (
	"context"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/wursta/gomigrator/internal/migrator"
)

type PgMigrator struct {
	dsn string
	mu  sync.Mutex
	db  *sqlx.DB
}

func New() *PgMigrator {
	return &PgMigrator{}
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

func (m *PgMigrator) Up(ctx context.Context, migrations []migrator.Migration) error {
	for i := range migrations {
		migration := migrations[i]
		migration.UpHandlerContext(ctx, nil)
	}
	return nil
}
