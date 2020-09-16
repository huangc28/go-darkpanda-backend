package utilcmd

import (
	"errors"
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/spf13/cobra"
)

var GenJwtTokenByUuid = &cobra.Command{
	Use:   "tuuid <uuid>",
	Short: "Generate JWT token by user uuid to request API.",
	Long:  "Generate JWT token by using user uuid and secret.",
	RunE:  GenJwtTokenByUuidFunc,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("exactly one user uuid must be specified to generate jwt.")
		}

		return nil
	},
}

func GenJwtTokenByUuidFunc(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	secret := config.GetAppConf().JwtSecret

	if SecretF != "" {
		secret = SecretF
	}

	jwtToken, err := jwtactor.CreateToken(uuid, secret)

	if err != nil {
		return err
	}

	fmt.Printf("generated token %s", jwtToken)

	return nil
}
