package postgres

import (
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/furadx/iam-go/internal/apiserver/store"
)

type datastore struct {
	db *gorm.DB
}

var (
	once     sync.Once
	instance store.Factory
)

// GetPostgresFactoryOr 创建 postgres 存储实例（单例）。
func GetPostgresFactoryOr(dsn string) (store.Factory, error) {
	var err error
	once.Do(func() {
		var db *gorm.DB
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			return
		}

		sqlDB, dbErr := db.DB()
		if dbErr != nil {
			err = dbErr
			return
		}

		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)

		instance = &datastore{db: db}
	})

	if instance == nil || err != nil {
		return nil, fmt.Errorf("failed to get postgres store: %w", err)
	}

	return instance, nil
}

func (ds *datastore) Users() store.UserStore {
	return newUsers(ds)
}

func (ds *datastore) Close() error {
	db, err := ds.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
