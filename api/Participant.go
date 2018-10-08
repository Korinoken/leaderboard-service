package api

type Participant struct {
	ParticipantId int    `json:"id"`
	TournamentId  int    `json:"tournament_id"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	FinalRank     int    `json:"final_rank"`
}
