// This package registers all dependencies share across the application includes DAO, services... etc
// Use interfaces in `contracts` package in pair with the dependency resolver to return appropriate proper
// abstract type that can be used in any package domain.

package deps

import (
	"sync"

	"github.com/golobby/container"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
)

type DepContainer struct {
	Container cintrnal.Container
}

type DepRegistrar func() error

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

func (dep *DepContainer) RegisterUserDAO() DepRegistrar {
	return func() error {
		dep.Container.Transient(func() contracts.UserDAOer {
			return user.NewUserDAO(db.GetDB())
		})

		return nil
	}
}

func (dep *DepContainer) Run() error {
	depRegistrars := []DepRegistrar{
		dep.RegisterUserDAO(),
	}

	for _, depRegistrar := range depRegistrars {
		if err := depRegistrar(); err != nil {
			return err
		}
	}

	return nil
}
