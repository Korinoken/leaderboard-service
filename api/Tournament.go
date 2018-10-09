package api

import (
	"time"
)

type Tournament struct {
	Id   int       `json:"id" gorm:"primary_key"`
	Name string    `json:"name"`
	Date time.Time `json:"started_at"`
	Url  string    `json:"url"`
}
