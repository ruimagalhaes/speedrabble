package main

import (
	"encoding/json"
	"io"
	"log"
	"main/game"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Read JSON data from a file (you can change this to read from any source)
	file, err := os.Open("static/words.json")
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

	var wordMap = make(map[string]int)
	// Unmarshal JSON data into the map
	if err := json.Unmarshal(byteValue, &wordMap); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	e := echo.New()
	e.Static("/frontend", "./frontend")
	e.Use(middleware.Logger())

	e.File("/", "frontend/index.html")
	e.File("/script.js", "frontend/script.js")
	e.File("/style.css", "frontend/style.css")

	gameHandler := game.GameHandler{
		CurrentGames: make(map[string]game.Game),
		WordMap:      wordMap,
		NPlayTiles:   7,
	}

	e.GET("/start", gameHandler.StartGame)
	e.POST("/guess", gameHandler.GuessWord)
	e.GET("/tiles", gameHandler.GetNewTiles)
	e.GET("/end", gameHandler.EndGame)

	e.Logger.Fatal(e.Start(":8080"))
}
