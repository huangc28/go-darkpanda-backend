package main

import utilcmd "github.com/huangc28/go-darkpanda-backend/internal/util_cmd"

// This package uses SQL parser from `sqlc` to generate go models in go code. All migration files reside in `db/migrations`. All migration files are prefixed with version number.
// genmodel package generate models from most update to date migration version. For instance, we got the most up to date version of migration from DB:
//
//   version: 7
//   dirty: false
//
// when we run:
//   go run cmd/genmodel/main.go gen
//
// we will collect content from migration files from 1 ~ 7 and cluter them in a master file `db/schema.sql` in which we will generate our go code from.

func main() {
	utilcmd.Execute()
}
