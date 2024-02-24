package db

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/juliotorresmoreno/tana-api/logger"
	"github.com/juliotorresmoreno/tana-api/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	logger_gorm "gorm.io/gorm/logger"
)

var applog = logger.SetupLogger()
var DefaultClient *gorm.DB
var DefaultCache *redis.Client

func Setup() {
	var err error
	DefaultClient, err = NewClient()
	if err == nil {
		applog.Info("Connected to database")
	} else {
		applog.Panic("Failed conection to database")
	}

	DefaultClient.AutoMigrate(&models.User{})
	DefaultClient.AutoMigrate(&models.Mmlu{})
	DefaultClient.AutoMigrate(&models.Credential{})

	DefaultCache, err = NewRedisClient()
	if err == nil {
		applog.Info("Connected to cache")
	} else {
		applog.Panic("Failed conection to cache")
	}
}

func NewClient() (*gorm.DB, error) {
	driver := os.Getenv("DATABASE_DRIVER")
	url := os.Getenv("DATABASE_URL")
	switch driver {
	case "postgres":
		return newPostgresClient(url, 10)
	}
	return &gorm.DB{}, errors.New("postgres isn't valid")
}

func newPostgresClient(dsn string, poolSize int) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger_gorm.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger_gorm.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger_gorm.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(poolSize)

	return db, nil
}
