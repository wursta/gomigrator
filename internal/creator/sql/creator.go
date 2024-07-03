package creatorsql

import (
	"os"
	"time"

	"github.com/wursta/gomigrator/internal/utils"
)

type SQLCreator struct {
	dir string
}

func New(dir string) *SQLCreator {
	return &SQLCreator{
		dir: dir,
	}
}

func (s *SQLCreator) Create(migrationName string) (*os.File, error) {
	f, err := utils.CreateMigrationFile(
		time.Now(),
		s.dir,
		migrationName,
		utils.GenerateRandomString(5),
		"sql",
	)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.WriteString(`-- migration: up

-- migration: down`)
	if err != nil {
		return nil, err
	}

	return f, nil
}
