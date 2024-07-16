![test](https://github.com/wursta/gomigrator/actions/workflows/test.yaml/badge.svg?branch=main)
![integration test](https://github.com/wursta/gomigrator/actions/workflows/integration-test.yaml/badge.svg?branch=main)

# Go Migrator
SQL migrations utility

## Usage

### Install
```
go get github.com/wursta/gomigrator && go install github.com/wursta/gomigrator/...@latest
```

### Run

#### Create migrations
```
gomigrator create create_some_table --migrations-dir=migrations
```

#### Up migrations
```
gomigrator up --migrations-dir=migrations --db-dsn=postgres://user:pass@dbhost:5432/dbname
```

#### Down migrations
```
gomigrator down --migrations-dir=migrations --db-dsn=postgres://user:pass@dbhost:5432/dbname
```

#### Redo last successfull migration
```
gomigrator redo --migrations-dir=migrations --db-dsn=postgres://user:pass@dbhost:5432/dbname
```

### Config

#### Config file
Create config file config.yaml

```
migrations_dir: "./migrations_up"
db_dsn: "postgres://test:test@localhost:5432/migrator_up_test"
```
Use `--config` flag in commands:
```
gomigrator up --config=./migrations_up
```

#### ENV-variables
You can use environment variables

```
GOMIGRATOR_MIGRATIONS_DIR="./migrations_up" GOMIGRATOR_DB_DSN="postgres://test:test@localhost:5432/migrator_up_test" gomigrator up
```

### Options
```
Usage:
  gomigrator [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Create a new migration file
  dbversion   Get current database version
  down        Rollback last success migration
  help        Help about any command
  redo        Redo last success migration
  status      Migrations status table
  up          Apply new migrations
  version     Show version

Flags:
      --config string           Config file (YAML format)
      --db-dsn string           Database connection in DSN format
  -h, --help                    help for gomigrator
      --migrations-dir string   Directory with migration files

```
