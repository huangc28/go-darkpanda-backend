package db

import (
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-txdb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

type Conn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowx(query string, args ...interface{}) *sqlx.Row
}

var dbInstance *sqlx.DB

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
		testdriver, err := sqlx.Open("txdb", dsn)

		if err != nil {
			log.Fatalf("Failed to connect to test db %s", err.Error())
		}

		dbInstance = testdriver

		log.Info("Database connected!")

		return
	}

	driver, err := sqlx.Open("postgres", dsn)

	if err != nil {
		log.WithFields(log.Fields{
			"db dsn": dsn,
		}).Fatalf("Failed to connect to database %s", err.Error())
	}

	dbInstance = driver

	log.Info("Database connected!")
}

func GetDB() *sqlx.DB {
	return dbInstance
}
