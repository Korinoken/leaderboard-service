package api

type TournamentResult struct {
	TournamentId   int    `json:"tournament_id" gorm:"-"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	FinalRank      int    `json:"final_rank"`
	TournamentName string `gorm:"-"`
}
