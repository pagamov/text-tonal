package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func makePostRequestToModel(c *gin.Context) Analyz {
	var jsonData []byte
	var err error
	var response *http.Response
	var body []byte

	var analyz Analyz

	jsonData = getJsonData(c)

	response, err = http.Post("http://localhost:8081/predict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error making POST request:", err)
	}
	defer response.Body.Close()

	body, err = io.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
	}

	err = json.Unmarshal(body, &analyz)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	return analyz
}

func getJsonData(c *gin.Context) []byte {
	var jsonData []byte
	var err error
	requestData := map[string]interface{}{
		"text": c.Query("text"),
	}

	jsonData, err = json.Marshal(requestData)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
	}
	return jsonData
}
