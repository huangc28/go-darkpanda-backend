package db

import (
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-txdb"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

var db *sql.DB

type DBConf struct {
	Host     string
	Port     uint
	User     string
	Password string
	Dbname   string
}

type TestDBConf struct {
	Host     string
	Port     uint
	User     string
	Password string
	Dbname   string
}

func InitDB(conf DBConf, testConf TestDBConf, isTestEnv bool) {
	log.Printf("is test %t", isTestEnv)

	// we need to recognize the running environment
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.Dbname,
	)

	if isTestEnv {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			testConf.Host,
			testConf.Port,
			testConf.User,
			testConf.Password,
			testConf.Dbname,
		)

		txdb.Register("txdb", "postgres", dsn)
		testdriver, err := sql.Open("txdb", dsn)

		if err != nil {
			log.Fatalf("Failed to connect to test db %s", err.Error())
		}

		db = testdriver

		log.Info("Database connected!")

		return
	}

	driver, err := sql.Open("postgres", dsn)

	if err != nil {
		log.WithFields(log.Fields{
			"db dsn": dsn,
		}).Fatalf("Failed to connect to database %s", err.Error())
	}

	db = driver

	log.Info("Database connected!")
}

func GetDB() *sql.DB {
	return db
}
