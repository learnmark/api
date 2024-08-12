package store

import (
	"gorm.io/gorm"
)

type GeneralDao struct {
	DB *gorm.DB
}

func NewGeneralDao(d *gorm.DB) *GeneralDao {
	return &GeneralDao{d}
}
