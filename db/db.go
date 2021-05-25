package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/DATA-DOG/go-txdb"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

type Conn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
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
	TimeZone string
}

type TestDBConf struct {
	Host     string
	Port     uint
	User     string
	Password string
	Dbname   string
	TimeZone string
}

func InitDB(conf DBConf, testConf TestDBConf, isTestEnv bool) {
	if isTestEnv {
		log.Printf("is test %v", testConf)

		initTestDB(testConf)

		log.Info("Test database connected!")
		return
	}

	// we need to recognize the running environment
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.Dbname,
		conf.TimeZone,
	)

	driver, err := sqlx.Open("postgres", dsn)

	if err != nil {
		log.WithFields(log.Fields{
			"db dsn": dsn,
		}).Fatalf("Failed to connect to database %s", err.Error())
	}

	dbInstance = driver
	dbInstance.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

	log.Info("DEV database connected!")
}

func initTestDB(testConf TestDBConf) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
		testConf.Host,
		testConf.Port,
		testConf.User,
		testConf.Password,
		testConf.Dbname,
		testConf.TimeZone,
	)

	txdb.Register("txdb", "postgres", dsn)
	testdriver, err := sqlx.Open("txdb", dsn)

	if err != nil {
		log.Fatalf("Failed to connect to test db %s", err.Error())
	}

	dbInstance = testdriver
	dbInstance.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

	log.Info("Database connected!")
}

func GetDB() *sqlx.DB {
	return dbInstance
}
