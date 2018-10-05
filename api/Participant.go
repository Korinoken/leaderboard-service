package api

type Participant struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	FinalRank int    `json:"final_rank"`
}
