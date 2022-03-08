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
	gameDetailURL = "https://www.nintendo.com/store/products/%s/"
	titleURL = "https://ec.nintendo.com/US/us/titles/%d"
)

type item struct {
	Slug       string  `json:"slug"`
	Discount   float32 `json:"eshop_discount_price"`
	Origin     float32 `json:"eshop_list_price_na"`
	SaleEnds   string  `json:"sale_ends"`
	Cover      string  `json:"cover_art_url"`
	Nsuid      int     `json:"nsuid_na"`
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
				if game, err := CreateGame(i.Slug, fmt.Sprintf(gameDetailURL, i.Slug)); err == nil {
					updatePrice(i, game)
				} else {
					if game, err := CreateGame(i.Slug, fmt.Sprintf(titleURL, i.Nsuid)); err == nil {
						updatePrice(i, game)
					} else {
						log.Error(fmt.Sprintf("Create game error: %s %v", i.Slug, err))
					}
				}
				<-ch
			}(i)
		}
	}
	wg.Wait()
}

func CreateGame(slug, url string) (Game, error) {
	log.Info("LoadURL: ", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		log.Error("Http Get Error: ", err)
		return Game{}, err
	}
	title := htmlquery.FindOne(doc, "//h1")
	if title == nil {
		return Game{}, errors.New("No Title!")
	}

	desc := htmlquery.FindOne(doc, "//div[starts-with(@class, 'RichTextstyles__Html')]")
	if desc == nil {
		return Game{}, errors.New("No Desc content!")
	}

	intro := strings.TrimSpace(strings.Split(strings.ReplaceAll(strings.ReplaceAll(
		htmlquery.InnerText(desc), "\n                ", " "), "\n\n", ""), "Read more")[0])

	list, err := htmlquery.QueryAll(doc, "//div[starts-with(@class, 'ProductInfostyles__InfoRow')]")
	if err != nil {
		panic(err)
	}

	date := ""
	HasChinese := false

	for _, item := range list {
		heading := htmlquery.InnerText(htmlquery.FindOne(
			item, "h3[starts-with(@class, 'Headingstyles__StyledH')]"))
		info := htmlquery.InnerText(htmlquery.FindOne(
			item, "div[starts-with(@class, 'ProductInfostyles__InfoDescr')]/div"))
		if heading == "Release date" {
			date = info
			break
		} else if heading == "Supported languages" {
			if strings.Contains(info, "Chinese") {
				HasChinese = true
			}
		}
	}

	if date == "" {
		return Game{}, errors.New("No release date!")
	}
	t, err := utils.ParseDate(date)
	if err != nil {
		return Game{}, err
	}

	game := Game{
		EnTitle: htmlquery.InnerText(title),
		Slug: slug,
		Desc: intro,
		HasChinese: HasChinese,
		ReleaseTime: datatypes.Date(t),
	}

	result := DB.Create(&game)

	if game.ID != 0 { // For test

		for _, item := range list {
			heading := htmlquery.InnerText(htmlquery.FindOne(
				item, "h3[starts-with(@class, 'Headingstyles__StyledH')]"))
			qc, err := htmlquery.QueryAll(item, "div[starts-with(@class, 'ProductInfostyles__InfoDescr')]//a")
			if err != nil {
				return Game{}, err
			}

			var data []string
			for _, i := range qc {
				data = append(data, htmlquery.InnerText(i))
			}
			if heading == "Genre" {
				BindGenres(game.ID, data)
			} else if heading == "Publisher" {
				BindPublishers(game.ID, data)
			} else if heading == "Developer" { // Maybe not used
				BindDevelopers(game.ID, data)
			}
		}

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
		if err = utils.WriteThumbImg(slug); err != nil {
			log.Info(err)
		} else {
			game.ThumbImg = fmt.Sprintf("%s.png", slug)
			DB.Save(&game)
		}
	} else {
		log.Error(result.Error)
	}
	return game, nil
}
