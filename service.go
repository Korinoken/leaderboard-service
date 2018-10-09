package main

import (
	_ "bufio"
	"fmt"
	"github.com/Korinoken/leaderboard-service/api"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	_ "os"
	"strconv"
)

type Config struct {
	BaseUrl     string
	ApiKey      string
	weights     map[int]int
	SvcHost     string
	DbName      string
	ServicePort int
}

type LeaderboardService struct {
}

func (c *LeaderboardService) Run(cfg Config) error {
	db, err := c.InitDB(cfg)
	if err != nil {
		return err
	}
	u, err := url.Parse(cfg.BaseUrl)
	if err != nil {
		log.Printf("Error while parsing url: %v", err)
	}
	model := LeaderboardModel{db: db,
		baseURL:    u,
		apiKey:     cfg.ApiKey,
		httpClient: http.DefaultClient,
		weights:    cfg.weights}

	ws := new(restful.WebService)
	ws.
		Path("/leaderboard").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.POST("{tournaments}").To(model.AddTournamentAndResults).
		Doc("add tournament and participants by url without domain").
		Operation("AddTournamentAndResults").Consumes())
	ws.Route(ws.DELETE("{tournaments}/{tournament-url}").To(model.DeleteTournament).
		Doc("delete tournament and results by url without domain").
		Operation("DeleteTournament"))
	ws.Route(ws.GET("{tournaments}/{tournament-url}").To(model.GetTournamentData).
		Doc("Get details for a tournament by url without domain").
		Operation("GetTournamentData"))
	ws.Route(ws.GET("scoreboard").To(model.GetScore).
		Doc("Get scores").
		Operation("GetScore"))
	ws.Route(ws.GET("participants/{participant-name}").To(model.GetParticipantData).
		Doc("Get details for participant").
		Operation("GetScore"))
	restful.Add(ws)

	//TODO: fix swagger and add annotations
	svcConfig := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: cfg.SvcHost,
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath: "/apidocs/",
		//SwaggerFilePath: "/Customers/emicklei/Projects/swagger-ui/dist"}
		SwaggerFilePath: "C:\\Go_projects\\pkg\\windows_amd64\\github.com\\emicklei\\swagger-ui-2.1.4\\dist"}
	swagger.InstallSwaggerService(svcConfig)

	log.Printf("start listening on localhost %v", cfg.SvcHost)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", cfg.ServicePort), nil))

	return nil
}
func ConnectDb(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", cfg.DbName)
	return db, err
}

func (c LeaderboardService) InitDB(cfg Config) (*gorm.DB, error) {
	db, err := ConnectDb(cfg)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&api.TournamentResult{})
	db.AutoMigrate(&api.Tournament{})
	return db, nil
}

func main() {
	cfg := &Config{}
	service := LeaderboardService{}
	cfgFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panicf("Cannot read config file: %v", err)
	}
	cfgFileData := gjson.Parse(string(cfgFile))
	cfgMap, ok := cfgFileData.Value().(map[string]interface{})
	if !ok {
		log.Panicf("Cannot parse config file")
	}

	mapstructure.Decode(cfgMap, &cfg)
	weightsMap := cfgFileData.Get("weights").Value()
	cfg.weights = make(map[int]int)
	for name, data := range weightsMap.(map[string]interface{}) {
		intKey, err := strconv.Atoi(name)
		if err != nil {
			log.Panicf("Error while parsing weights data: %v", err)
		}
		cfg.weights[intKey] = int(data.(float64))
	}

	log.Printf("Loaded config: %+v", cfg)
	service.Run(*cfg)
}
