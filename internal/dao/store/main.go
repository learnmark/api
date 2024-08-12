package store

import (
	"fmt"
	"sync"

	"github.com/learnmark/learnmark/internal/dao"
	"github.com/learnmark/learnmark/internal/model"
	"github.com/learnmark/learnmark/pkg/db"
	"gorm.io/gorm"
)

type Dao struct {
	DB *gorm.DB
}

func (d *Dao) GeneralDao() dao.GeneralDao {
	return NewGeneralDao(d.DB)
}

func (d *Dao) UserDao() dao.UserDao {
	return NewUserDao(d.DB)
}

type ColumnType interface {
	DatabaseTypeName() string // varchar
}

func GetDao(opts *db.Options) (dao.Interface, error) {
	var daoInterface dao.Interface
	var once sync.Once

	if opts == nil {
		return nil, fmt.Errorf("failed to get database options")
	}

	var err error
	var dbIns *gorm.DB
	once.Do(func() {
		options := &db.Options{
			Driver:                opts.Driver,
			Host:                  opts.Host,
			Port:                  opts.Port,
			Username:              opts.Username,
			Password:              opts.Password,
			Database:              opts.Database,
			MaxIdleConnections:    opts.MaxIdleConnections,
			MaxOpenConnections:    opts.MaxOpenConnections,
			MaxConnectionLifeTime: opts.MaxConnectionLifeTime,
			Logger:                opts.Logger,
		}
		dbIns, err = db.NewGORM(options)
	})

	initErr := dbIns.AutoMigrate(
		&model.User{},
	)

	if initErr != nil {
		return nil, fmt.Errorf("failed to init learnmark database: %w", err)
	}

	daoInterface = &Dao{dbIns}

	if err != nil {
		return nil, fmt.Errorf("failed to get learnmark dao: %+v, error: %w", daoInterface, err)
	}

	return daoInterface, nil
}
