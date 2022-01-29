package main

import (
	"fmt"

	"github.com/dongweiming/eshop-prices/config"
	"github.com/dongweiming/eshop-prices/models"

	"github.com/dongweiming/go-eshop/api"
)

func main() {
	db := models.Initialize()
	fmt.Println(db)
	fmt.Println(api.GetAllGames("US"))
}
