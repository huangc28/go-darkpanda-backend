package genmodel

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/huangc28/go-darkpanda-backend/config"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func init() {
	config.InitConfig()
}

func GetMigrationInfo(db *sql.DB) (int, bool, error) {
	var (
		version int
		dirty   bool
	)

	err := db.QueryRow(`
		SELECT version, dirty
		FROM schema_migrations
	`).Scan(&version, &dirty)

	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

// pick list of file names that are <= specific version.
func pickMigrationsByVersion(files []os.FileInfo, version int) []os.FileInfo {
	var suitedFiles []os.FileInfo

	for _, file := range files {
		migVer, err := strconv.Atoi(strings.Split(file.Name(), "_")[0])

		if err != nil {
			log.Printf("Failed to parse version number of %s, skipping...", file.Name())

			continue
		}

		sufSegs := strings.Split(file.Name(), ".")
		migType := sufSegs[len(sufSegs)-2 : len(sufSegs)-1][0]

		if migVer <= version && migType == "up" {
			suitedFiles = append(suitedFiles, file)
		}
	}

	return suitedFiles
}

func appendFileContentToDestFile(files []os.FileInfo, src string, dest string) {
	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer destFile.Close()

	if err != nil {
		log.Fatalf("Failed to open / create dest file %s", err.Error())
	}

	for _, file := range files {
		func(file os.FileInfo) {
			cByte, err := ioutil.ReadFile(filepath.Join(src, file.Name()))

			if err != nil {
				log.Fatalf("Failed to read bytes from src file %s", err.Error())
			}

			if _, err := destFile.Write(cByte); err != nil {
				log.Fatalf("Failed when piping content from %s, exiting... with error %s", filepath.Join(src, file.Name()), err.Error())
			}
		}(file)
	}
}

func Gen(cmd *cobra.Command, args []string) error {
	env := os.Getenv("ENV")

	log.Printf("DEBUG 1 %v", env)

	// read latest migration info from database
	appConf := config.GetAppConf()

	psqlDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		appConf.DBConf.Host,
		appConf.DBConf.Port,
		appConf.DBConf.User,
		appConf.DBConf.Password,
		appConf.DBConf.Dbname,
	)

	if env == "test" {
		testDBConf := config.GetTestDBConf()

		psqlDSN = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			testDBConf.Host,
			testDBConf.Port,
			testDBConf.User,
			testDBConf.Password,
			testDBConf.Dbname,
		)
	}

	db, err := sql.Open("postgres", psqlDSN)

	if err != nil {
		log.Fatalf("Failed to connect to psql %v", err.Error())
		return err
	}

	// ---------- read migration info of the project  ----------
	// read version info from schema_migrations table
	version, dirty, err := GetMigrationInfo(db)

	log.Printf("\n\n version: %d\n dirty: %t", version, dirty)

	if err != nil {
		log.Fatalf("Failed to read migration info %s", err.Error())

		return err
	}

	if dirty {
		log.Fatal("Migration seems dirty! Please fix the migration first")

		return err
	}

	// ---------- read `migration up` files from migration directory ----------
	cwd, _ := os.Getwd()
	dirPath := filepath.Join(cwd, "db/migrations")

	files, err := ioutil.ReadDir(dirPath)

	if err != nil {
		log.Fatalf("failed to read migrations %s, %s", dirPath, err.Error())

		return err
	}

	mFiles := pickMigrationsByVersion(files, version)

	appendFileContentToDestFile(mFiles, dirPath, filepath.Join(cwd, "db/schema.sql"))

	// ---------- execute sqlc generate command ----------
	var out bytes.Buffer
	osCmd := exec.Command("sqlc", "generate")
	osCmd.Env = append(os.Environ())
	osCmd.Stdout = &out

	if err := osCmd.Run(); err != nil {
		log.Fatalf("Failed to run 'sqlc generate' command %s", err.Error())

		return err
	}

	fmt.Printf("go models generated successfully from %s %s\n", dirPath, out.String())

	return nil
}

var genModelCmd = &cobra.Command{
	Use:   "gen",
	Short: "read / collect SQL from list of migrations files to genderate models in go code.",
	RunE:  Gen,
}
