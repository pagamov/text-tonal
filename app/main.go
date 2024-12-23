package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	initRedis()

	var router Router
	router.Init()
	router.AddMethod()
	router.Start("8080")
}

// can be multiple labels for one word
type Info struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// for word we got N info marks for each label
type Word struct {
	Word  string `json:"Word"`
	Label string `json:"Label"`
}

type Analyz struct {
	Label string `json:"Label"`
	Words []Word `json:"Words"`
}

type Statistics struct {
	Date  string `json:"date"`
	Text  string `json:"text"`
	Label string `json:"label"`
	Words []Word `json:"words"`
}

type Router struct {
	router *gin.Engine
}

func (api *Router) Init() {
	// gin.SetMode(gin.ReleaseMode)
	api.router = gin.Default()
}

func (api *Router) AddMethod() {
	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics/:begin/:end", statistics)
	api.router.GET("/ping", ping)
	api.router.GET("/logs", getLogs)
}

func (api *Router) Start(port string) {
	api.router.Run(fmt.Sprintf(":%s", port))
}

func getLogs(c *gin.Context) {
	var statistics []Statistics = getLog("01.01.2000", "01.01.2030")
	c.IndentedJSON(http.StatusOK, statistics)
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func analyze(c *gin.Context) {
	// 	POST API/analyze?text=some text to parse
	var jsonData []byte
	var analyz Analyz
	jsonData = getJsonData(c)
	fmt.Println(string(jsonData))

	if checkIfInRedis(c) {
		analyz = getFromRedis(jsonData)
		log.Println("Get from Redis")
	} else {
		analyz = makePostRequestToModel(c)
		addToRedis(jsonData, analyz)
		log.Println("Added to Redis")
	}

	jsonAnalyz, err := json.Marshal(analyz)
	if err != nil {
		fmt.Println(err)
		return
	}
	addLog(c.Query("text"), analyz.Label, string(jsonAnalyz))
	c.IndentedJSON(http.StatusOK, analyz)
}

func validateDate(dateStr string) bool {
	layout := "02.01.2006" // dd.mm.yyyy
	_, err := time.Parse(layout, dateStr)
	return err == nil
}

func parseDate(date string) time.Time {
	time, _ := time.Parse("02.01.2006", date)
	return time
}

func statistics(c *gin.Context) {
	// GET API/statistics/dd.mm.yyyy/dd.mm.yyyy

	var data_from string
	var data_to string
	var exist bool

	var from time.Time
	var to time.Time

	if data_from, exist = c.Params.Get("begin"); !exist || !validateDate(data_from) {
		c.IndentedJSON(http.StatusBadRequest, "data_from not Exist or not dd.mm.yyyy")
		return
	}

	if data_to, exist = c.Params.Get("end"); !exist || !validateDate(data_to) {
		c.IndentedJSON(http.StatusBadRequest, "data_end not Exist or not dd.mm.yyyy")
		return
	}

	from = parseDate(data_from)
	to = parseDate(data_to)
	if !from.Before(to) {
		c.IndentedJSON(http.StatusBadRequest, "data_end before data_start")
		return
	}

	var res []Statistics = getLog(data_from, data_to)

	// var res []Statistics = []Statistics{
	// 	{
	// 		Date:  "01/01/1977 14:20:00",
	// 		Text:  "Some text",
	// 		Count: 10,
	// 		Label: "label",
	// 		Words: []Word{
	// 			{
	// 				Word:  "word",
	// 				Label: "label",
	// 			},
	// 		},
	// 	},
	// }

	c.IndentedJSON(http.StatusOK, res)
}
