package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/huangc28/go-darkpanda-backend/config"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func init() {
	config.InitConfig()
}

//  - run the latest migrations
//  - record SQL that is going to be executed for historical traceback in the future
func main() {
	ac := config.GetAppConf()

	pwd, _ := os.Getwd()

	sourceUrl := fmt.Sprintf("file://%s", filepath.Join(pwd, "db/migrations"))

	m, err := migrate.New(
		sourceUrl,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			ac.DBConf.User,
			ac.DBConf.Password,
			ac.DBConf.Host,
			ac.DBConf.Port,
			ac.DBConf.Dbname,
		),
	)

	if err != nil {
		log.Fatalf("failed to initialize go migrate instance %s", err.Error())
	}

	log.WithFields(log.Fields{
		"database url": fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			ac.DBConf.User,
			ac.DBConf.Password,
			ac.DBConf.Host,
			ac.DBConf.Port,
			ac.DBConf.Dbname,
		),
	}).Info("database connected!")

	if err = m.Up(); err != nil {
		log.Fatalf("failed to run migrations %s", err.Error())
	}
}
