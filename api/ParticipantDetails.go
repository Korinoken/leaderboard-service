package api

type ParticipantDetails struct {
	UserName    string `gorm:"primary_key"`
	DisplayName string
	Country     string
}
