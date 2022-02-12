package models

import (
	"fmt"
	"errors"
	"strings"

	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
)

var ids []int

type BaseModel struct {
	gorm.Model
	CnName string  `gorm:"type:varchar(200);default:"`
	EnName string  `gorm:"type:varchar(200);uniqueIndex`
}

type Publisher struct {
	BaseModel
}

type Developer struct {
	BaseModel
}

type Genre struct {
	BaseModel
}

type BaseGameModel struct {
	gorm.Model
	GameID int32 `gorm:"uniqueIndex`
	TargetID int32
}

type GamePublisher struct {
	BaseGameModel
}

type GameDeveloper struct {
	BaseGameModel
}

type GameGenre struct {
	BaseGameModel
}

func GetPublishers(id int32) (items []Publisher) {
	DB.Select("id").Where(&GamePublisher{BaseGameModel{TargetID: id}}).Find(&ids)
	if len(ids) == 0 {
		return  nil
	}

	DB.Find(&items, ids)
	return
}

func GetGenres(id int32) (items []Genre) {
	DB.Select("id").Where(&GameGenre{BaseGameModel{TargetID: id}}).Find(&ids)
	if len(ids) == 0 {
		return  nil
	}

	DB.Find(&items, ids)
	return
}

func GetDevelopers(id int32) (items []Developer) {
	DB.Select("id").Where(&GameGenre{BaseGameModel{TargetID: id}}).Find(&ids)
	if len(ids) == 0 {
		return  nil
	}

	DB.Find(&items, ids)

	return
}

func BindPublishers(gid int32, items []string) {
	var titles []string
	for _, i := range items {
		titles = append(titles, strings.Title(i))
	}
	if items == nil {
		return
	}
	var publishers []Publisher
	DB.Where("en_name IN ?", titles).Find(&publishers)

	if len(publishers) != len(items) {
		pubMap := make(map[string]bool)
		for _, p := range publishers {
			pubMap[p.EnName] = true
		}
		var pubs []Publisher
		for _, t := range titles {
			if _, ok := pubMap[t]; !ok {
				pubs = append(pubs, Publisher{BaseModel{EnName: t}})
			}
		}
		DB.Create(&pubs)
		publishers = append(publishers, pubs...)
	}

	for _, pub := range publishers {
		var p GamePublisher
		if err := DB.Where("game_id = ? AND target_id= ?", gid, pub.ID).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				DB.Create(&GamePublisher{BaseGameModel{GameID: gid, TargetID: int32(pub.ID)}})
			} else {
				log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
			}
		}
	}

}

func BindGenres(gid int32, items []string) {
	var titles []string
	for _, i := range items {
		titles = append(titles, strings.Title(i))
	}
	if items == nil {
		return
	}
	var genres []Genre
	DB.Where("en_name IN ?", titles).Find(&genres)

	if len(genres) != len(items) {
		pubMap := make(map[string]bool)
		for _, p := range genres {
			pubMap[p.EnName] = true
		}
		var pubs []Genre
		for _, t := range titles {
			if _, ok := pubMap[t]; !ok {
				pubs = append(pubs, Genre{BaseModel{EnName: t}})
			}
		}
		DB.Create(&pubs)
		genres = append(genres, pubs...)
	}

	for _, pub := range genres {
		var p GameGenre
		if err := DB.Where("game_id = ? AND target_id= ?", gid, pub.ID).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				DB.Create(&GameGenre{BaseGameModel{GameID: gid, TargetID: int32(pub.ID)}})
			} else {
				log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
			}
		}
	}

}

func BindDevelopers(gid int32, items []string) {
	var titles []string
	for _, i := range items {
		titles = append(titles, strings.Title(i))
	}
	if items == nil {
		return
	}
	var developers []Developer
	DB.Where("en_name IN ?", titles).Find(&developers)

	if len(developers) != len(items) {
		pubMap := make(map[string]bool)
		for _, p := range developers {
			pubMap[p.EnName] = true
		}
		var pubs []Developer
		for _, t := range titles {
			if _, ok := pubMap[t]; !ok {
				pubs = append(pubs, Developer{BaseModel{EnName: t}})
			}
		}
		DB.Create(&pubs)
		developers = append(developers, pubs...)
	}

	for _, pub := range developers {
		var p GameDeveloper
		if err := DB.Where("game_id = ? AND target_id= ?", gid, pub.ID).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				DB.Create(&GameDeveloper{BaseGameModel{GameID: gid, TargetID: int32(pub.ID)}})
			} else {
				log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
			}
		}
	}
}

//type Platforms struct {
//BaseModel
//}
