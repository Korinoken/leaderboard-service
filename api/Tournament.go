package api

import "time"

type Tournament struct {
	Id   int       `json:"id"`
	Name string    `json:"name"`
	Date time.Time `json:"started_at"`
	Url  string    `json:"full_challonge_url"`
}
