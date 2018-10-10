package api

type ParticipantDetails struct {
	Name  string
	Score int
	Games int
}
type ByScore []ParticipantDetails

func (b ByScore) Len() int           { return len(b) }
func (b ByScore) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByScore) Less(i, j int) bool { return b[i].Score > b[j].Score }
