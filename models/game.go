package models

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/datatypes"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"

	"github.com/dongweiming/eshop-prices/config"
)

var once sync.Once

type database struct {
	instance    *gorm.DB
}

var db *database

// Aliases,Publisher,Developer,Platforms,Genres

type Game struct {
	gorm.Model
	CnTitle string
	EnTitle string
	Score float32
	MetacriticScore uint8
	ThumbImg string
	Slug string
	Kind uint8
	Desc string
	ReleaseTime datatypes.Date
}

func Initialize() *gorm.DB {
	once.Do(func() {
		db = new(database)
		conf := config.ReadConfig()
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
			conf.DB.User,
			conf.DB.Passwd,
			conf.DB.Host,
			conf.DB.Port,
			conf.DB.Database)
		db_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

		db_.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Game{})

		if err != nil {
			panic(err)
		}
		db.instance = db_
	})
	return db.instance
}
