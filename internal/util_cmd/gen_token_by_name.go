package utilcmd

import (
	"errors"
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/spf13/cobra"
)

var GenJwtTokenByName = &cobra.Command{
	Use:   "tname <name>",
	Short: "Generate JWT token by user uuid to request API.",
	Long:  "Generate JWT token by using user uuid and secret.",
	RunE:  GenJwtTokenByNameFunc,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("exactly one user uuid must be specified to generate jwt.")
		}

		return nil
	},
}

func init() {
	// initialize database
	config.InitConfig()
	dbConf := config.GetDBConf()
	db.InitDB(db.DBConf{
		Host:     dbConf.Host,
		Port:     dbConf.Port,
		User:     dbConf.User,
		Password: dbConf.Password,
		Dbname:   dbConf.Dbname,
	}, false)
}

func GenJwtTokenByNameFunc(cmd *cobra.Command, args []string) error {
	// retrieve user uuid by username
	var (
		uuid   string
		secret string = config.GetAppConf().JwtSecret
	)

	username := args[0]

	if SecretF != "" {
		secret = SecretF
	}

	db := db.GetDB()
	if err := db.
		QueryRow("SELECT uuid FROM users WHERE username = $1", username).
		Scan(&uuid); err != nil {
		return err
	}

	jwtToken, err := jwtactor.CreateToken(uuid, secret)

	if err != nil {
		return err
	}

	fmt.Printf("generated token %s", jwtToken)

	return nil
}
