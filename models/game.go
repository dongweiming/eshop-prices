package models

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
var DB *gorm.DB

// Aliases,Publisher,Developer,Platforms,Genres

type Game struct {
	Id      int32
	CnTitle string             `gorm:"default:"`
	EnTitle string             `gorm:"type:varchar(200);index:,default:"`
	Score float32              `gorm:"default:10.0"`
	MetacriticScore uint8      `gorm:"default:10"`
	ThumbImg string            `gorm:"default:"`
	Slug string                `gorm:"primaryKey:,type:varchar(100);uniqueIndex"`
	Kind uint8                 `gorm:"default:0"`
	Desc string
	ReleaseTime datatypes.Date
}

func Initialize() *gorm.DB {
	once.Do(func() {
		db = new(database)
		conf := config.ReadConfig()
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
			conf.DB.User,
			conf.DB.Passwd,
			conf.DB.Host,
			conf.DB.Port,
			conf.DB.Database)
		db_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})

		err = db_.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Game{})
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		db.instance = db_
	})
	return db.instance
}

func init() {
	DB = Initialize()
}

func GetAllGames() []Game {
	var games []Game
	DB.Find(&games)
	return games
}
