package grpc

import (
	"context"
	golog "log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	v1 "github.com/learnmark/learnmark/api/learnmark/v1"
	"github.com/learnmark/learnmark/internal/dao"
	"github.com/learnmark/learnmark/internal/dao/store"
	"github.com/learnmark/learnmark/internal/mid"
	"github.com/learnmark/learnmark/internal/model"
	servicev1 "github.com/learnmark/learnmark/internal/service/v1"
	"github.com/learnmark/learnmark/pkg/db"
	"github.com/learnmark/learnmark/pkg/log"
	"github.com/learnmark/learnmark/pkg/utils"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/gorm/logger"
)

func Run(ctx context.Context, opts utils.Options) error {
	//init grpc server and run
	l, err := net.Listen(opts.Network, opts.GRPCAddr)
	if err != nil {
		return err
	}
	go func() {
		defer func() error {
			if err := l.Close(); err != nil {
				return err
			}
			return nil
		}()
		<-ctx.Done()
	}()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(mid.Auth), selector.MatchFunc(mid.AllButHealthZ)),
		),
		grpc.ChainStreamInterceptor(
			selector.StreamServerInterceptor(auth.StreamServerInterceptor(mid.Auth), selector.MatchFunc(mid.AllButHealthZ)),
		),
	)

	var daoInterface dao.Interface
	if daoInterface, err = initDao(); err != nil {
		return err
	}

	learnmarkService := servicev1.NewlearnmarkService(daoInterface)

	v1.RegisterLearnmarkServer(s, learnmarkService)

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()

	go func() error {
		log.L(ctx).Infof("grpc listen on: %s\n", opts.GRPCAddr)
		if err := s.Serve(l); err != nil {
			return err
		}
		return nil
	}()

	return nil
}

func initDao() (dao.Interface, error) {
	newLogger := logger.New(
		golog.New(os.Stdout, "", golog.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.LogLevel(viper.GetInt("data.database.log-level")),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	options := db.Options{
		Driver:                viper.GetString("data.database.driver"),
		Host:                  viper.GetString("data.database.host"),
		Port:                  viper.GetString("data.database.port"),
		Username:              viper.GetString("data.database.user"),
		Password:              viper.GetString("data.database.password"),
		Database:              viper.GetString("data.database.database"),
		MaxIdleConnections:    viper.GetInt("data.database.max-idle-connections"),
		MaxOpenConnections:    viper.GetInt("data.database.max-open-connections"),
		AutoCreateAdmin:       viper.GetBool("data.database.auto-create-admin"),
		MaxConnectionLifeTime: time.Duration(viper.GetInt("data.database.max-connection-lifetime")) * time.Second,
		Logger:                newLogger,
	}
	learnmarkDao, err := store.GetDao(&options)
	if err != nil {
		return nil, err
	}

	if options.AutoCreateAdmin == true {
		// get create user or not config from config.yaml
		admin, err := learnmarkDao.UserDao().GetByName(model.User{
			Name: "admin",
		})
		if err != nil {
			return nil, err
		}
		if admin.Id == uuid.Nil {
			var newAdmin model.User
			newAdmin.Id = uuid.New()
			newAdmin.Name = "admin"
			newAdmin.Email = "admin@learnmark.io"
			newAdmin.Password = "admin"
			newAdmin.IsSuperAdmin = true
			_, err = learnmarkDao.UserDao().Create(newAdmin)
			if err != nil {
				return nil, err
			}
		}
	}

	return learnmarkDao, nil
}
