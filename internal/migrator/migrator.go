package migrator

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type MigrationHandlerContext func(context.Context, *sqlx.Tx) error

type Migration struct {
	Name               string
	FilePath           string
	UpHandlerContext   MigrationHandlerContext
	DownHandlerContext MigrationHandlerContext
}
