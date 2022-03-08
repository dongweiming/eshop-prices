package main

import (
	"fmt"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/datatypes"
	"github.com/araddon/dateparse"
	log "github.com/sirupsen/logrus"

	"github.com/dongweiming/eshop-prices/models"
	"github.com/dongweiming/go-eshop/api"
	"github.com/dongweiming/eshop-prices/utils"
)


func main() {
	db := models.DB
	items := api.GetAllGames("US")

	// 展示不用事务提交方式
	// tx := db.Begin()
	// defer func() {
	//if r := recover(); r != nil {
	//		tx.Rollback()
	//	}
	// }()
	for _,  item := range items {
		t, err := dateparse.ParseAny(item.ReleaseDate)
		if err != nil {
			log.Error(fmt.Sprintf("Parse data error: %v\n", err))
			continue
		}
		game := models.Game{
			EnTitle: strings.Replace(item.Title, "™", "", -1),
			Slug: item.Slug,
			Desc: item.Desc,
			ReleaseTime: datatypes.Date(t),
		}

		// TODO: clause.OnConflict is not work
		var g models.Game
		var id int32
		if err := db.Where("slug = ?", game.Slug).First(&g).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				db.Create(&game)
				id = game.ID
			} else {
				log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
			}
		} else {
			db.Model(&g).Where("slug = ?", game.Slug).Updates(game)
			id = g.ID

		}

		if g.ThumbImg == "" {
			if err = utils.WriteThumbImg(item.Slug); err != nil {
				log.Info(err)
			} else {
				g.ThumbImg = fmt.Sprintf("%s.png", item.Slug)
				db.Save(&g)
			}
		}

		models.BindGenres(id, item.Genres)
		models.BindPublishers(id, strings.Split(item.Publishers, ","))
		models.BindDevelopers(id, strings.Split(item.Developers, ","))
	}

	// tx.Commit()

	log.Info("Game's Data had Updated!")
}
