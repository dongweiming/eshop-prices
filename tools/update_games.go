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
 )

 func main() {
	db := models.DB
	items := api.GetAllGames("US")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var games []models.Game
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
		games = append(games, game)
	}

	// TODO: clause.OnConflict is not work
	for _, game := range games {
		var g models.Game
		if err := db.Where("slug = ?", game.Slug).First(&g).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				db.Create(&game)
			} else {
				log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
			}
		} else {
			db.Model(&game).Where("slug = ?", game.Slug).Updates(game)
		}
	}

	tx.Commit()

	log.Info("Game's Data had Updated!")
}
