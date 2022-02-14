package main

import (
	"os"
	"fmt"
	"sync"
	"strings"
	"errors"
	"io/ioutil"
	"net/http"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/datatypes"
	log "github.com/sirupsen/logrus"
	"github.com/antchfx/htmlquery"

	"github.com/dongweiming/go-eshop/eshop"
	"github.com/dongweiming/eshop-prices/utils"
	. "github.com/dongweiming/eshop-prices/models"
)

const (
	limit = 1000
	url = "https://www.nsgreviews.com/search/s?search=&offset=%d&limit=1000"
	gameDetailURL = "https://www.nintendo.com/games/detail/%s/"
)

type item struct {
	Slug       string  `json:"slug"`
	Discount   float32 `json:"eshop_discount_price"`
	Origin     float32 `json:"eshop_list_price_na"`
	SaleEnds   string  `json:"sale_ends"`
	Cover      string  `json:"cover_art_url"`
}

type Response struct {
	Total int          `json:"total"`
	Items []item       `json:"rows"`
}

func GetNsgData(page int) Response {
	log.Info(fmt.Sprintf("Fetch Nsg Data Page(%d) ...", page))
	resp, err := http.Get(fmt.Sprintf(url, page * limit))
	defer resp.Body.Close()
	if err != nil {
		log.Error("Http Get Error", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Read Data Error", err)
		os.Exit(1)
	}
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	fmt.Printf("%v", result)
	return result
}

func main() {
	var gameMap = make(map[string]Game)
	games := GetAllGames()
	for _, game := range games {
		gameMap[game.Slug] = game
	}
	page := 0

	var wg sync.WaitGroup
	ch := make(chan struct{}, 3)
	rs := GetNsgData(page)
	go UpdatePrice(gameMap, rs.Items)
	for page = 1; page < rs.Total / limit; page++ {
		ch <- struct{}{}
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			rs = GetNsgData(page)
			UpdatePrice(gameMap, rs.Items)
			<-ch
		}(page)
	}
	wg.Wait()

	log.Info("Game's Price Data had Updated!")

}

func updatePrice(i item, game Game) {
	var p Price

	t, _ := utils.ParseDate(i.SaleEnds)

	price := Price{
		GID: game.ID,
		Discount: i.Discount,
		Origin: i.Origin,
		SaleEnds: datatypes.Date(t),
		Country: eshop.US,
	}
	if err := DB.Where("g_id = ?", game.ID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			DB.Create(&price)
		} else {
			log.Error(fmt.Sprintf("Upsert data error: %v\n", err))
		}
	} else {
		DB.Model(&p).Where("g_id = ?", game.ID).Updates(price)
	}
}

func UpdatePrice(m map[string]Game, items []item) {
	print(len(items))
	var wg sync.WaitGroup
	ch := make(chan struct{}, 3)
	for _, i := range items {
		if game, ok := m[i.Slug]; ok {
			updatePrice(i, game)
		} else {
			ch <- struct{}{}
			wg.Add(1)
			go func(i item) {
				defer wg.Done()
				if game, err := CreateGame(i.Slug); err != nil {
					updatePrice(i, game)
				}
				<-ch
			}(i)
		}
	}
	wg.Wait()
}

func CreateGame(slug string) (Game, error) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf(gameDetailURL, slug))
	if err != nil {
		log.Error("Http Get Error", err)
		return Game{}, err
	}
	title := htmlquery.FindOne(doc, "//h1[contains(@class, 'game-title')]")
	if title == nil {
		return Game{}, errors.New("No Title!")
	}

	desc := htmlquery.FindOne(doc, "//div[@class='overview-content']")
	if desc == nil {
		return Game{}, errors.New("No Desc content!")
	}

	intro := strings.TrimSpace(strings.Split(strings.ReplaceAll(strings.ReplaceAll(
		htmlquery.InnerText(desc), "\n                ", " "), "\n\n", ""), "Read more")[0])

	date := htmlquery.FindOne(doc, "//div[contains(@class, 'release-date')]/dd")

	if date == nil {
		return Game{}, errors.New("No release date!")
	}
	t, err := utils.ParseDate(htmlquery.InnerText(date))
	if err != nil {
		return Game{}, err
	}

	game := Game{
		EnTitle: htmlquery.InnerText(title),
		Slug: slug,
		Desc: intro,
	ReleaseTime: datatypes.Date(t),
	}

	result := DB.Create(&game)

	if game.ID != 0 { // For test
		item := htmlquery.FindOne(doc, "//div[contains(@class, 'genre')]/dd")
		if item != nil {
			BindGenres(game.ID, strings.Split(htmlquery.InnerText(item), ","))
		}

		item = htmlquery.FindOne(doc, "//div[contains(@class, 'developer')]/dd")
		if item != nil {
			BindDevelopers(game.ID, strings.Split(htmlquery.InnerText(item), ","))
		}

		item = htmlquery.FindOne(doc, "//div[contains(@class, 'genre publisher')]/dd")
		if item != nil {
			BindPublishers(game.ID, strings.Split(htmlquery.InnerText(item), ","))
		}
	} else {
		log.Error(result.Error)
	}
	return game, nil
}
