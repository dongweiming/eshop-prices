package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/dongweiming/eshop-prices/models"
	"github.com/dongweiming/go-eshop/api"
 )

func main() {
	var gameMap = make(map[string]models.Game)
	games := models.GetAllGames()
	for _, game := range games {
		gameMap[game.EnTitle] = game
	}
	items := api.GetMetacriticItems()
	db := models.DB
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range items {
		if game, ok := gameMap[item.Title]; ok {
			game.MetacriticScore = uint8(item.Score)
			tx.Save(&game)
		}
	}
	tx.Commit()

	log.Info("Game's Metacritic Score Data had Updated!")
}
