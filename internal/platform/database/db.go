package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Engine *gorm.DB

func Init(host string, port int, db string, user, password string) {
	var err error
	var dsn string

	dbConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info,
				Colorful:      false,
			}),
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, db)
	Engine, err = gorm.Open(mysql.Open(dsn), dbConfig)
	if err != nil {
		panic(err)
	}
	t, err := Engine.DB()
	if err != nil {
		panic(err)
	}
	t.SetConnMaxLifetime(60 * time.Second)
	t.SetMaxIdleConns(10)
	t.SetMaxOpenConns(512)
}
