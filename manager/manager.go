package manager

import (
	"context"
	"flag"
	"strings"
	"sync"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	log "github.com/sirupsen/logrus"
)

var _manager *Manager

type initializer func() error

type Manager struct {
	initials     map[string]initializer
	initialState map[string]bool
	initialNames []string
	initComplete bool
	ctx          context.Context
	sync.RWMutex
}

func (m *Manager) IsInited(name string) bool {
	return m.initialState[name]
}

// Exec immediately run initialize after register
func (m *Manager) Exec(name string, init initializer) error {
	m.Lock()
	defer m.Unlock()

	m.initialNames = append(m.initialNames, name)
	m.initials[name] = init
	init()
	m.initialState[name] = true

	return nil
}

func (m *Manager) Register(name string, init initializer) error {
	m.Lock()
	defer m.Unlock()
	m.initialNames = append(m.initialNames, name)
	m.initials[name] = init
	return nil
}

func (m *Manager) ExecAppConfig() *Manager {
	m.Exec("init app config", func() error {
		config.InitConfig()

		return nil
	})

	return m
}

func (m *Manager) ExecDBInit() *Manager {
	m.Exec("init DB", func() error {
		conf := config.GetAppConf()

		db.InitDB(
			db.DBConf{
				Host:     conf.PGHost,
				Port:     conf.PGPort,
				User:     conf.PGUser,
				Password: conf.PGPassword,
				Dbname:   conf.PGDbname,
				TimeZone: conf.PGTimeZone,
			},
			db.TestDBConf{
				Host:     conf.TestPGHost,
				Port:     conf.TestPGPort,
				User:     conf.TestPGUser,
				Password: conf.TestPGPassword,
				Dbname:   conf.TestPGDbname,
				TimeZone: conf.TestPGTimeZone,
			},
			flag.Lookup("test.v") != nil,
		)

		return nil
	})

	return m
}

func (m *Manager) ExecRedisInit() *Manager {
	m.Exec("init redis", func() error {

		appConf := config.GetAppConf()

		err := db.InitRedis(db.RedisConf{
			Addr:     appConf.RedisHost,
			Password: appConf.RedisPassword,
			DB:       int(appConf.RedisDb),
		})

		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	return m
}

func (m *Manager) ExecFireStoreInit() *Manager {
	err := darkfirestore.InitFireStore(
		m.ctx,
		darkfirestore.InitOptions{
			CredentialFile: config.GetAppConf().FirestoreCredentialFile,
		},
	)

	if err != nil {
		log.Fatalf("Failed to initialize firestore client: \n%s", err.Error())
	}

	return m
}

func (m *Manager) Initialize() error {
	log.Debugf("initializers: %s", strings.Join(m.initialNames, ", "))

	for _, name := range m.initialNames {
		if !m.IsInited(name) {
			initFunc := m.initials[name]

			// TODO: we should pass request context to callback
			err := initFunc()

			if err != nil {
				return err
			}

			m.initialState[name] = true
		}
	}

	m.initComplete = true
	return nil
}

func (m *Manager) Run(f func()) {
	if !m.initComplete {
		err := m.Initialize()
		if err != nil {
			log.Errorf("oops.. app is not completely initialized: %s", err.Error())
			return
		}
	}

	f()
}

func NewManager(ctx context.Context) *Manager {
	_manager := &Manager{
		initials:     make(map[string]initializer),
		initialState: make(map[string]bool),
		ctx:          ctx,
	}

	return _manager
}

func NewDefaultManager(ctx context.Context) *Manager {
	_manager := NewManager(ctx)
	// we need to initialize log register before call log
	_manager.
		ExecAppConfig().
		ExecDBInit().
		ExecRedisInit().
		ExecFireStoreInit()

	return _manager
}

func GetManager() *Manager {
	return _manager
}
