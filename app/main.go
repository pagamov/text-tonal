package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jbrukh/bayesian"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func main() {

	classifier := bayesian.NewClassifier(Good, Bad)
	goodStuff := []string{"tall", "rich", "handsome"}
	badStuff := []string{"tall", "poor", "smelly", "ugly"}
	classifier.Learn(goodStuff, Good)
	classifier.Learn(badStuff, Bad)

	scores, likely, _ := classifier.LogScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(scores, likely)

	probs, likely, _ := classifier.ProbScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(probs, likely)

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.Run(":8080")

}
