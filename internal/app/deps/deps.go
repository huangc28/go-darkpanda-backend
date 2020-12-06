// This package registers all dependencies share across the application includes DAO, services... etc
// Use interfaces in `contracts` package in pair with the dependency resolver to return appropriate proper
// abstract type that can be used in any package domain.

package deps

import (
	"sync"

	"github.com/golobby/container"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
	"github.com/huangc28/go-darkpanda-backend/internal/app/service"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
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

func (dep *DepContainer) Run() error {
	depRegistrars := []DepRegistrar{
		user.UserDaoServiceProvider(dep.Container),
		service.ServiceDAOServiceProvider(dep.Container),
		inquiry.InquiryDaoServiceProvider(dep.Container),
		chat.ChatDaoServiceProvider(dep.Container),
		payment.PaymentDAOServiceProvider(dep.Container),
		image.ImageDAOServiceProvider(dep.Container),
	}

	for _, depRegistrar := range depRegistrars {
		if err := depRegistrar(); err != nil {
			return err
		}
	}

	return nil
}
