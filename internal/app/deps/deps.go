// This package registers all dependencies share across the application includes DAO, services... etc
// Use interfaces in `contracts` package in pair with the dependency resolver to return appropriate proper
// abstract type that can be used in any package domain.

package deps

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/golobby/container"
	cinternal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	bankAccount "github.com/huangc28/go-darkpanda-backend/internal/app/bank_account"
	"github.com/huangc28/go-darkpanda-backend/internal/app/block"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/coin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/pubsuber"
	"github.com/huangc28/go-darkpanda-backend/internal/app/rate"
	"github.com/huangc28/go-darkpanda-backend/internal/app/register"

	gcsenhancer "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/gcs_enhancer"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
	"github.com/huangc28/go-darkpanda-backend/internal/app/service"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"google.golang.org/api/option"
)

type DepContainer struct {
	Container cinternal.Container
}

type DepRegistrar func() error
type ServiceProvider func(cinternal.Container) DepRegistrar

var (
	_depContainer *DepContainer
	once          sync.Once
)

func Get() *DepContainer {
	once.Do(func() {
		_depContainer = &DepContainer{
			Container: container.NewContainer(),
		}
	})

	return _depContainer
}

func (dep *DepContainer) TwilioServiceProvider(c cinternal.Container) DepRegistrar {
	return func() error {
		c.Transient(func() twilio.TwilioServicer {
			appConf := config.GetAppConf()

			return twilio.New(twilio.TwilioConf{
				AccountSID:   appConf.TwilioAccountID,
				AccountToken: appConf.TwilioAuthToken,
			})
		})

		return nil
	}
}

func (dep *DepContainer) DarkFirestoreServiceProvider(c cinternal.Container) DepRegistrar {
	return func() error {
		ctx := context.Background()

		err := darkfirestore.InitFireStore(
			ctx,
			darkfirestore.InitOptions{
				CredentialFile: config.GetAppConf().FirestoreCredentialFile,
			},
		)

		if err != nil {
			log.Fatalf("failed to initialize darkfirestore %v", err)
		}

		c.Singleton(func() darkfirestore.DarkFireStorer {
			return darkfirestore.Get()
		})

		return nil
	}
}

func (dep *DepContainer) GcsEnhancerServiceProvider(c cinternal.Container) DepRegistrar {
	return func() error {
		ctx := context.Background()

		client, err := storage.NewClient(
			ctx,
			option.WithCredentialsFile(
				fmt.Sprintf(
					"%s/%s",
					config.GetProjRootPath(),
					config.GetAppConf().GcsGoogleServiceAccountName,
				),
			),
		)

		if err != nil {
			log.Fatalf("failed to initialize google cloud storage %v", err)
		}

		enhancer := gcsenhancer.NewGCSEnhancer(
			client,
			config.GetAppConf().GcsBucketName,
		)

		c.Singleton(func() gcsenhancer.GCSEnhancerInterface {
			return enhancer
		})

		return nil
	}
}

func (dep *DepContainer) PubsuberServiceProvider(c cinternal.Container) DepRegistrar {
	return func() error {

		ctx := context.Background()

		log.Printf("DEBUG spot 1")

		// Setup google service account path.
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fmt.Sprintf("%s/%s", config.GetProjRootPath(), config.GetAppConf().FirestoreCredentialFile))
		client, err := pubsub.NewClient(ctx, config.GetAppConf().GcpProjectID)

		log.Printf("DEBUG spot 2")
		if err != nil {
			log.Fatal("failed to initialise google pubsub client")
		}

		c.Singleton(func() pubsuber.DPPubsuber {
			return pubsuber.New(client)
		})

		return nil
	}
}

func (dep *DepContainer) Run() error {
	depRegistrars := []DepRegistrar{
		dep.TwilioServiceProvider(dep.Container),
		dep.DarkFirestoreServiceProvider(dep.Container),
		dep.GcsEnhancerServiceProvider(dep.Container),
		dep.PubsuberServiceProvider(dep.Container),

		user.UserDaoServiceProvider(dep.Container),
		service.ServiceDAOServiceProvider(dep.Container),
		service.ServiceFSMProvider(dep.Container),
		inquiry.InquiryDaoServiceProvider(dep.Container),

		chat.ChatDaoServiceProvider(dep.Container),
		chat.ChatServiceServiceProvider(dep.Container),

		payment.PaymentDAOServiceProvider(dep.Container),
		image.ImageDAOServiceProvider(dep.Container),
		bankAccount.BankAccountDAOServiceProvider(dep.Container),
		auth.AuthDaoerServiceProvider(dep.Container),

		coin.CoinDAOServiceProvider(dep.Container),
		coin.CoinPackageDaoServiceProvider(dep.Container),
		coin.UserBalanceDAOServiceProvider(dep.Container),

		rate.RateDAOServiceProvider(dep.Container),
		register.RegisterDaoServiceProvider(dep.Container),

		block.BlockDAOServiceProvider(dep.Container),
	}

	for _, depRegistrar := range depRegistrars {
		if err := depRegistrar(); err != nil {
			return err
		}
	}

	return nil
}
