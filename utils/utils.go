package utils

import (
	"fmt"
	"unicode"

	"gorm.io/datatypes"
	"github.com/araddon/dateparse"
	log "github.com/sirupsen/logrus"
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
		log.Error(fmt.Sprintf("Parse data error: %v\n", err))
		return datatypes.Date{}, err
	}
	return datatypes.Date(t), nil
}
