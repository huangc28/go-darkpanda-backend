// This package registers all dependencies share across the application includes DAO, services... etc
// Use interfaces in `contracts` package in pair with the dependency resolver to return appropriate proper
// abstract type that can be used in any package domain.

package deps

import (
	"context"
	"log"
	"sync"

	"github.com/golobby/container"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	bankAccount "github.com/huangc28/go-darkpanda-backend/internal/app/bank_account"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
	"github.com/huangc28/go-darkpanda-backend/internal/app/service"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"github.com/spf13/viper"
)

type DepContainer struct {
	Container cintrnal.Container
}

type DepRegistrar func() error
type ServiceProvider func(cintrnal.Container) DepRegistrar

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

func (dep *DepContainer) TwilioServiceProvider(c cintrnal.Container) DepRegistrar {
	return func() error {
		c.Transient(func() twilio.TwilioServicer {
			return twilio.New(twilio.TwilioConf{
				AccountSID:   viper.GetString("twilio.account_id"),
				AccountToken: viper.GetString("twilio.auth_token"),
			})
		})

		return nil
	}
}

func (dep *DepContainer) DarkFirestoreServiceProvider(c cintrnal.Container) DepRegistrar {
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

func (dep *DepContainer) Run() error {
	depRegistrars := []DepRegistrar{
		dep.TwilioServiceProvider(dep.Container),
		dep.DarkFirestoreServiceProvider(dep.Container),
		user.UserDaoServiceProvider(dep.Container),
		service.ServiceDAOServiceProvider(dep.Container),
		inquiry.InquiryDaoServiceProvider(dep.Container),

		chat.ChatDaoServiceProvider(dep.Container),
		chat.ChatServiceServiceProvider(dep.Container),

		payment.PaymentDAOServiceProvider(dep.Container),
		image.ImageDAOServiceProvider(dep.Container),
		bankAccount.BankAccountDAOServiceProvider(dep.Container),
	}

	for _, depRegistrar := range depRegistrars {
		if err := depRegistrar(); err != nil {
			return err
		}
	}

	return nil
}
