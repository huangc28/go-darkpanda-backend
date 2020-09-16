package manager

import (
	"flag"
	"strings"
	"sync"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	log "github.com/sirupsen/logrus"
)

type initializer func() error

type Manager struct {
	initials     map[string]initializer
	initialState map[string]bool
	initialNames []string
	initComplete bool
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

		log.Printf("app conf %v", conf.DBConf)

		db.InitDB(db.DBConf{
			Host:     conf.DBConf.Host,
			Port:     conf.DBConf.Port,
			User:     conf.DBConf.User,
			Password: conf.DBConf.Password,
			Dbname:   conf.DBConf.Dbname,
		}, flag.Lookup("test.v") != nil)

		return nil
	})

	return m
}

func (m *Manager) ExecRedisInit() *Manager {
	rdsConf := config.GetAppConf().RedisConf

	err := db.InitRedis(db.RedisConf{
		Addr:     rdsConf.Addr,
		Password: rdsConf.Password,
		DB:       rdsConf.DB,
	})

	if err != nil {
		log.Fatal(err)
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

func NewManager() *Manager {
	manager := &Manager{
		initials:     make(map[string]initializer),
		initialState: make(map[string]bool),
	}

	return manager
}

func NewDefaultManager() *Manager {
	manager := &Manager{
		initials:     make(map[string]initializer),
		initialState: make(map[string]bool),
	}

	// we need to initialize log register before call log
	manager.
		ExecAppConfig().
		ExecDBInit().
		ExecRedisInit()

	// In the future, we can use condition build to make
	// swagger not be compiled in the production environment.
	//if config.GetAppEnv() == config.STAGING || config.GetAppEnv() == config.DEVELOPMENT {
	//swagger.GenerateDoc()
	//}

	//manager.RegisterMySQL("init db(mysql)")

	//manager.Register("init redis", func() error {
	//err := database.InitRedisConnections()
	//if err != nil {
	//return err
	//}
	//return nil
	//})

	return manager
}
