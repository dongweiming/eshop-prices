package models

import (
	"gorm.io/datatypes"
)

type Price struct {
	ID         int32              `gorm:"primaryKey;autoIncrement"`
	GID        int32              `gorm:"index:"`
	Discount   float32
	Origin     float32
	Country    int32
	SaleEnds   datatypes.Date     `gorm:"default:"`
}
