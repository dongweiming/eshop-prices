package main

import (
	"fmt"
	"time"
	"strings"

	"github.com/gocolly/colly"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/dongweiming/eshop-prices/models"
	"github.com/dongweiming/eshop-prices/utils"
)

const (
	url = "https://zh.wikipedia.org/wiki/%d年任天堂Switch游戏列表"
	start_year = 2017
)

type item struct {
	Titles []string
	HasChinese bool
}

func main() {
	end_year := time.Now().Year()
	var gameMap = make(map[string]models.Game)
	games := models.GetAllGames()
	for _, game := range games {
		gameMap[game.EnTitle] = game
	}
	db := models.DB
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var items []*item

	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(models.UA),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: models.PARALLEL})

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		i := &item{}
		eles := e.DOM.Find("td")
		eles.Eq(0).Contents().Each(func(_ int, n *goquery.Selection) {
			if n.Is("br") {
				return
			}
			text := strings.TrimSpace(strings.Split(n.Text(), "(")[0])
			text = strings.Split(text, "（")[0]

			if text != "" && !strings.HasPrefix(text, "＊") && !strings.HasPrefix(text, "*") {
				i.Titles = append(i.Titles, text)
			}
		})
		i.HasChinese = eles.Eq(11).Text() == "是"
		items = append(items, i)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error: Request ", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	for p := start_year; p <= end_year; p++ {
		c.Visit(fmt.Sprintf(url, p))
	}

	c.Wait()

	for _, item := range items {
		for _, title := range item.Titles {
			if game, ok := gameMap[title]; ok {
				var titles []string
				for _, t := range item.Titles {
					if t != title {
						titles = append(titles, t)
					}
				}
				if utils.IsChinese(item.Titles[0]) {
					game.CnTitle = titles[0]
					game.Aliases = strings.Join(titles[1:], ",")
				} else {
					game.Aliases = strings.Join(titles, ",")
				}
				game.HasChinese = item.HasChinese
				tx.Save(&game)
			}
		}
	}
	tx.Commit()

	log.Info("Game's Title Data had Updated!")
}
