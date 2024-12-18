package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
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
	Word string `json:"word"`
	Info []Info `json:"info"`
}

type Analyz struct {
	Count int64  `json:"count"`
	Label string `json:"label"`
	Words []Word `json:"words"`
}

type Statistics struct {
	Date  string `json:"date"`
	Text  string `json:"text"`
	Count int64  `json:"count"`
	Label string `json:"label"`
	Words []Word `json:"words"`
}

type Router struct {
	router *gin.Engine
}

func (api *Router) Init() {
	api.router = gin.Default()
}

func (api *Router) AddMethod() {
	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics", statistics)
	api.router.GET("/ping", ping)
}

func (api *Router) Start(port string) {
	api.router.Run(fmt.Sprintf(":%s", port))
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func analyze(c *gin.Context) {
	// 	POST API/analyze?text=some text to parse
	// 	RES =  {
	//         "count" : "Number of words : Int64",
	//         "label" : "soft max label of text : String",
	//         "words" : [
	//             {
	//                 "word" : "word itself : String",
	//                 "info" : [
	//                     {
	//                         "label" : "some label from learning labels : String",
	//                         "value" : "percentage : Int8"
	//                     }
	//                 ]
	//             }
	//         ]
	// }

	res := Analyz{
		Count: 10,
		Label: "label",
		Words: []Word{
			{
				Word: "word",
				Info: []Info{
					{Label: "label", Value: 10},
				},
			},
		},
	}

	c.IndentedJSON(http.StatusOK, res)
}

func statistics(c *gin.Context) {
	// GET API/statistics?date_begin=“dd.mm.yyyy”&date_end==“dd.mm.yyyy”
	// RES =  [{
	// 	"date" : "date of request : Date",
	// 	"text" : "text : String",
	// 	"count" : "Number of words : Int64",
	// 			"label" : "soft max label of text : String",
	// 			"words" : [
	// 				{  "word" : "word itself : String",
	// 					"info" : [{
	// 							"label" : "some label from learning labels : String",
	// 							"value" : "percentage : Int8"
	// 						}]
	// 				}
	// 			]
	// 	}]

	var res []Statistics = []Statistics{
		{
			Date:  "01/01/1977 14:20:00",
			Text:  "Some text",
			Count: 10,
			Label: "label",
			Words: []Word{
				{
					Word: "word",
					Info: []Info{
						{
							Label: "label",
							Value: 0,
						},
					},
				},
			},
		},
	}

	c.IndentedJSON(http.StatusOK, res)
}
