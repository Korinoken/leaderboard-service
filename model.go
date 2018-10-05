package main

import (
	"./api"
	"net/http"

	"encoding/json"
	//"github.com/emicklei/go-restful"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/url"
)

type LeaderboardModel struct {
	BaseURL    *url.URL
	ApiKey     string
	Db         *gorm.DB
	UserAgent  string
	httpClient *http.Client
}

func (l LeaderboardModel) getParticipantsAndStandings(tournamentName string) []api.Participant {
	rel := &url.URL{Path: "tournaments/" + tournamentName + "/participants"}
	requestUrl := l.BaseURL.ResolveReference(rel)
	requestQuery := requestUrl.Query()
	requestQuery.Add("api_key", l.ApiKey)
	requestUrl.RawQuery = requestQuery.Encode()
	request, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		log.Panicf("Error while creating request:%v", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", l.UserAgent)

	response, err := l.httpClient.Do(request)
	if err != nil {
		log.Panicf("Error while doing request:%v", err)
	}
	defer response.Body.Close()
	log.Printf("Call complete for tournament %v :%v", tournamentName, response.Status)

	respData, err := ioutil.ReadAll(response.Body)

	var respArray []map[string]api.Participant
	err = json.Unmarshal(respData, &respArray)
	if err != nil {
		log.Panicf("Error while parsing Json:%v", err)
	}
	var participants []api.Participant
	for _, participant := range respArray {
		participants = append(participants, participant["participant"])
	}
	return participants
}
