package main

import (
	"encoding/json"
	"github.com/Korinoken/leaderboard-service/api"
	"github.com/emicklei/go-restful"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
)

type LeaderboardModel struct {
	baseURL    *url.URL
	apiKey     string
	db         *gorm.DB
	userAgent  string
	httpClient *http.Client
	weights    map[int]int
}

func (l LeaderboardModel) DeleteTournament(request *restful.Request, response *restful.Response) {
	tournamentName := request.PathParameter("tournament-url")
	currentTournament := api.Tournament{}
	l.db.Where("url = ?", tournamentName).First(&currentTournament).Delete(api.Tournament{})
	l.db.Where("tournament_id = ?", currentTournament.Id).Delete(api.TournamentResult{})
	response.WriteHeader(http.StatusOK)
}

func (l LeaderboardModel) AddTournamentAndResults(request *restful.Request, response *restful.Response) {
	req := &api.AddTournamentRequest{}
	err := request.ReadEntity(req)
	if err != nil {
		log.Printf("Cannot read tournament name:%v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	tournamentDetails := l.getTournamentDetails(req.TournamentName)
	l.db.Create(tournamentDetails)
	log.Printf("Added tournament details for tournament%v", req.TournamentName)
	participants := l.getParticipantsAndStandings(req.TournamentName)
	for _, participant := range participants {
		participant.TournamentId = tournamentDetails.Id
		l.db.Create(participant)
	}
	log.Printf("Added %v participants for tournamnt %v", len(participants), req.TournamentName)
	response.WriteHeader(http.StatusOK)
}

func (l LeaderboardModel) GetScore(request *restful.Request, response *restful.Response) {
	calculatedScores, err := l.calculateScore()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Created scoreboard")
	var details []api.ParticipantDetails
	err = l.db.Find(&details).Error
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while fetching details %v", err)
		return
	}
	detailMap := map[string]api.ParticipantDetails{}

	for _, val := range details {
		detailMap[val.UserName] = val
	}
	var results []api.ParticipantResults
	for _, result := range *calculatedScores {
		currentDetail := detailMap[result.ChallongeName]
		result.Name = currentDetail.DisplayName
		result.Country = currentDetail.Country
		results = append(results, result)
	}
	response.WriteEntity(results)
}
func (l LeaderboardModel) CreateParticipantDetails(request *restful.Request, response *restful.Response) {
	req := &api.ParticipantDetails{}
	err := request.ReadEntity(req)
	if err != nil {
		log.Printf("Cannot read participant data:%v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	l.db.Assign(req).FirstOrCreate(req)
	response.WriteHeader(http.StatusOK)
}

func (l LeaderboardModel) GetParticipantDetails(request *restful.Request, response *restful.Response) {
	resp := &api.ParticipantDetails{}
	participantName := request.PathParameter("participant-name")
	if l.db.Where("user_name = ?", participantName).First(resp).RecordNotFound() {
		log.Printf("Cannot find details for participant %v", participantName)
		response.WriteHeader(http.StatusNotFound)
	} else {
		log.Printf("Fetched details for participant %v", participantName)
		response.WriteEntity(resp)
	}
}

func (l LeaderboardModel) GetParticipantData(request *restful.Request, response *restful.Response) {
	var participantEntries []api.TournamentResult
	participantName := request.PathParameter("participant-name")
	err := l.db.Table("tournament_results").
		Select("tournament_id,tournament_results.name as name,username,final_rank,tournaments.name as tournament_name").
		Joins("left join tournaments on tournament_id = id").
		Where("username = ?", participantName).Scan(&participantEntries).Error
	if err != nil {
		log.Printf("Error while accessing participant list: %v", err)
		response.WriteHeader(http.StatusNotFound)
		return
	}
	if len(participantEntries) > 0 {
		log.Printf("Fetched details for player %v", participantName)
		response.WriteEntity(&participantEntries)
	} else {
		log.Printf("Player not found %v", participantName)
		response.WriteHeader(http.StatusNotFound)
	}
}

func (l LeaderboardModel) GetTournamentData(request *restful.Request, response *restful.Response) {
	var tournamentDetails api.Tournament
	tourneyName := request.PathParameter("tournament-url")
	err := l.db.Where("url = ?", tourneyName).First(&tournamentDetails).Error
	if err != nil {
		switch err.Error() {
		case gorm.ErrRecordNotFound.Error():
			log.Printf("No details found for tournament %v", err)
			response.WriteHeader(http.StatusNotFound)
		default:
			log.Printf("Error while fetching tournament details %v", err)
			response.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	log.Printf("Fetched tournament data for :%v", tourneyName)
	response.WriteEntity(&tournamentDetails)
}

func (l LeaderboardModel) getDataFromApi(apiPath string) []byte {
	rel := &url.URL{Path: apiPath}
	requestUrl := l.baseURL.ResolveReference(rel)
	requestQuery := requestUrl.Query()
	requestQuery.Add("api_key", l.apiKey)
	requestUrl.RawQuery = requestQuery.Encode()
	request, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		log.Panicf("Error while creating request:%v", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", l.userAgent)

	response, err := l.httpClient.Do(request)
	if err != nil {
		log.Panicf("Error while doing request:%v", err)
	}
	defer response.Body.Close()
	log.Printf("Call complete for tournament %v :%v", apiPath, response.Status)

	respData, err := ioutil.ReadAll(response.Body)
	return respData
}

func (l LeaderboardModel) getParticipantsAndStandings(tournamentName string) []api.TournamentResult {
	respData := l.getDataFromApi("tournaments/" + tournamentName + "/participants")

	var respArray []map[string]api.TournamentResult
	err := json.Unmarshal(respData, &respArray)
	if err != nil {
		log.Panicf("Error while parsing Json:%v", err)
	}
	var participants []api.TournamentResult
	for _, participant := range respArray {
		participants = append(participants, participant["participant"])
	}
	return participants
}

func (l LeaderboardModel) getTournamentDetails(tournamentName string) api.Tournament {
	respData := l.getDataFromApi("tournaments/" + tournamentName)
	var tournamentDetails map[string]api.Tournament
	err := json.Unmarshal(respData, &tournamentDetails)
	if err != nil {
		log.Panicf("Error while parsing Json:%v", err)
	}
	return tournamentDetails["tournament"]
}

func (l LeaderboardModel) calculateScore() (*[]api.ParticipantResults, error) {
	var participants []api.TournamentResult
	err := l.db.Find(&participants).Error
	if err != nil {
		log.Printf("Error while accessing users: %v", err)
		return nil, err
	}

	userScores := make(map[string]*api.ParticipantResults)
	for _, participant := range participants {
		if len(participant.Username) > 0 {
			if _, exists := userScores[participant.Username]; exists {
				userScores[participant.Username].Score += l.weights[participant.FinalRank]
				userScores[participant.Username].Score += l.weights[99]
				userScores[participant.Username].Games += 1
			} else {
				userScores[participant.Username] = &api.ParticipantResults{ChallongeName: participant.Username,
					Score: l.weights[participant.FinalRank] + l.weights[99],
					Games: 1}
			}
		}
	}
	var resultArray []api.ParticipantResults
	for _, score := range userScores {
		score.Score = score.Score / (score.Games + 1)
		resultArray = append(resultArray, *score)
	}
	sort.Sort(api.ByScore(resultArray))
	return &resultArray, nil
}
