package migrator

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type MigrationHandlerContext func(context.Context, *sqlx.Tx) error

type Migration struct {
	ID                 *uint           `db:"id"`
	FilePath           string          `db:"file_path"`
	Name               string          `db:"file_name"`
	Status             MigrationStatus `db:"status"`
	CreateDT           time.Time       `db:"create_dt"`
	MigrateDT          *time.Time      `db:"migrate_dt"`
	UpHandlerContext   MigrationHandlerContext
	DownHandlerContext MigrationHandlerContext
}

type MigrationStatus string

const (
	MigrationStatusUnknown   MigrationStatus = "unknown"
	MigrationStatusNew       MigrationStatus = "new"
	MigrationStatusMigrating MigrationStatus = "migrating"
	MigrationStatusMigrated  MigrationStatus = "migrated"
	MigrationStatusFailed    MigrationStatus = "failed"
)

type MigrationDirection string

const (
	MigrationDirectionUp   MigrationDirection = "up"
	MigrationDirectionDown MigrationDirection = "down"
)
