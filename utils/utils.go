package utils

import (
	"os"
	"io"
	"fmt"
	"errors"
	"runtime"
	"unicode"
	"net/http"
	"path/filepath"

	"gorm.io/datatypes"
	"github.com/araddon/dateparse"
	log "github.com/sirupsen/logrus"
)

const (
	imageURL = "https://assets.nintendo.com/image/upload/ncom/en_US/games/switch/%s/hero"
)

func IsChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

func ParseDate(s string) (datatypes.Date, error) {
	t, err := dateparse.ParseAny(s)
	if err != nil {
		//log.Error(fmt.Sprintf("Parse data error: %v\n", err))
		return datatypes.Date{}, err
	}
	return datatypes.Date(t), nil
}

func WriteThumbImg(slug string) error {
	_, fileName, _, _ := runtime.Caller(0)
	assetsURL := filepath.Join(filepath.Dir(fileName), "../static/assets/%s.png")
	url := fmt.Sprintf(imageURL, fmt.Sprintf("%s/%s", string(slug[0]), slug))
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Info(fmt.Sprintf("Fetching: %s", url))
		return errors.New("Received non 200 response code")
	}

	file, err := os.Create(fmt.Sprintf(assetsURL, slug))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
