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

type Game struct {
	ID      int32              `gorm:"primaryKey;autoIncrement"`
	CnTitle string             `gorm:"default:"`
	EnTitle string             `gorm:"type:varchar(200);index:,default:"`
	Score float32              `gorm:"default:0.0"`
	MetacriticScore uint8      `gorm:"default:0"`
	ThumbImg string            `gorm:"type:varchar(100);default:"`
	Slug string                `gorm:"type:varchar(200);uniqueIndex"`
	Kind uint8                 `gorm:"default:0"`
	Desc string
	Aliases string
	HasChinese bool
	ReleaseTime datatypes.Date
}

func (g *Game) Publishers() []Publisher {
	return GetPublishers(g.ID)
}

func (g *Game) Developers() []Developer {
	return GetDevelopers(g.ID)
}

func (g *Game) Genres() []Genre {
	return GetGenres(g.ID)
}

//func (g *Game) Platforms() {
//
//}

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
