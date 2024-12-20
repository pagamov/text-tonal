package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	ctx    context.Context
	client *redis.Client
)

func initRedis() {
	ctx = context.Background()
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Test the connection
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}
	fmt.Println("Connected to Redis:", pong)
}

func checkIfInRedis(c *gin.Context) bool {
	// log.Println("check if in redis")
	// log.Println(string(getJsonData(c)))

	_, err := client.Get(ctx, string(getJsonData(c))).Result()
	if err == redis.Nil {
		return false
	} else if err != nil {
		log.Println(err)
	} else {
		return true
	}
	return false
}

func addToRedis(jsonData []byte, analyz Analyz) {
	resp, err := json.Marshal(analyz)
	if err != nil {
		log.Println("addToRedis() Marshal", err)
	}
	_, err = client.Set(ctx, string(jsonData), resp, 0).Result()
	if err != nil {
		log.Println("addToRedis()", err)
	}
}

func getFromRedis(jsonData []byte) Analyz {
	var analyz Analyz
	resp, err := client.Get(ctx, string(jsonData)).Bytes()
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(resp, &analyz)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}
	return analyz
}
