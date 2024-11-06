package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/rand"
)

type GuessRequest struct {
	Game  string `json:"game"`
	Guess string `json:"guess"`
}

type GameResponse struct {
	GameId string `json:"gameId"`
	Tiles  []Tile `json:"tiles"`
	Points int    `json:"points"`
}

type GoodGuessResponse struct {
	GameId string `json:"gameId"`
	Tiles  []Tile `json:"tiles"`
	Points int    `json:"points"`
	Valid  bool   `json:"valid"`
}

type BadGuessResponse struct {
	Valid bool `json:"valid"`
}

type Tile struct {
	Letter rune
	Points int
}

type Game struct {
	Id        string
	Words     []string
	BagTiles  []Tile
	UsedTiles []Tile
	PlayTiles []Tile
	Points    int
}

func (g *Game) PrintGame() {
	fmt.Println("===================================")
	fmt.Printf("GAME %s:\n", g.Id)
	bagStr := "BAT TILES"
	bagStr = bagStr + fmt.Sprintf(" [%d]: ", len(g.BagTiles))
	for _, tile := range g.BagTiles {
		bagStr = bagStr + string(tile.Letter) + " "
	}
	fmt.Println(bagStr)
	usedStr := "USED TILES:"
	usedStr = usedStr + fmt.Sprintf(" [%d]: ", len(g.UsedTiles))
	for _, tile := range g.UsedTiles {
		usedStr = usedStr + string(tile.Letter) + " "
	}
	fmt.Println(usedStr)
	playTiles := "PLAY TILES:"
	playTiles = playTiles + fmt.Sprintf(" [%d]: ", len(g.PlayTiles))
	for _, tile := range g.PlayTiles {
		playTiles = playTiles + string(tile.Letter) + " "
	}
	fmt.Println(playTiles)
	fmt.Println("POINTS: " + fmt.Sprint(g.Points))
	wordsStr := "WORDS: "
	for _, word := range g.Words {
		wordsStr = wordsStr + word + " "
	}
	fmt.Println(wordsStr)
	fmt.Printf("TOTAL TILES: %d\n", len(g.BagTiles)+len(g.UsedTiles)+len(g.PlayTiles))
	fmt.Println("===================================")
}

var currentGames = make(map[string]Game)
var wordMap = make(map[string]int)

var gamesMutex sync.Mutex
var nPlayTiles = 7
var scrabbleTiles = []Tile{
	// 1 point letters
	{'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1}, {'E', 1},
	{'A', 1}, {'A', 1}, {'A', 1}, {'A', 1}, {'A', 1}, {'A', 1}, {'A', 1}, {'A', 1}, {'A', 1},
	{'I', 1}, {'I', 1}, {'I', 1}, {'I', 1}, {'I', 1}, {'I', 1}, {'I', 1}, {'I', 1}, {'I', 1},
	{'O', 1}, {'O', 1}, {'O', 1}, {'O', 1}, {'O', 1}, {'O', 1}, {'O', 1}, {'O', 1},
	{'N', 1}, {'N', 1}, {'N', 1}, {'N', 1}, {'N', 1}, {'N', 1},
	{'R', 1}, {'R', 1}, {'R', 1}, {'R', 1}, {'R', 1}, {'R', 1},
	{'T', 1}, {'T', 1}, {'T', 1}, {'T', 1}, {'T', 1}, {'T', 1},
	{'L', 1}, {'L', 1}, {'L', 1}, {'L', 1},
	{'S', 1}, {'S', 1}, {'S', 1}, {'S', 1},
	{'U', 1}, {'U', 1}, {'U', 1}, {'U', 1},
	// 2 points letters
	{'D', 2}, {'D', 2}, {'D', 2}, {'D', 2},
	{'G', 2}, {'G', 2}, {'G', 2},
	// 3 points letters
	{'B', 3}, {'B', 3},
	{'C', 3}, {'C', 3},
	{'M', 3}, {'M', 3},
	{'P', 3}, {'P', 3},
	// 4 points letters
	{'F', 4}, {'F', 4},
	{'H', 4}, {'H', 4},
	{'V', 4}, {'V', 4},
	{'W', 4}, {'W', 4},
	{'Y', 4}, {'Y', 4},
	// 5 points letter
	{'K', 5},
	// 8 points letters
	{'J', 8},
	{'X', 8},
	// 10 points letters
	{'Q', 10},
	{'Z', 10},
}

const (
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength = 8
)

var seededRand *rand.Rand = rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

// TODO: change to generate unique and sequenctial IDs
func generateGameID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

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

	// Unmarshal JSON data into the map
	if err := json.Unmarshal(byteValue, &wordMap); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	e := echo.New()
	e.Static("/frontend", "./frontend")
	e.Use(middleware.Logger())

	// ADD GAME AS A MIDDLEWARE:
	// e.Use(databaseMiddleware(database))

	e.File("/", "frontend/index.html")
	e.File("/script.js", "frontend/script.js")
	e.File("/style.css", "frontend/style.css")

	// articlesHandler := handler.ArticlesHandler{}
	e.GET("/start", startGame)
	e.POST("/guess", guessWord)
	e.GET("/tiles", getNewTiles)
	e.GET("/end", endGame)

	e.Logger.Fatal(e.Start(":8080"))
}

func startGame(c echo.Context) error {
	gameTiles := make([]Tile, len(scrabbleTiles))
	copy(gameTiles, scrabbleTiles)
	shuffleTiles(&gameTiles)
	gameID := generateGameID()
	newGame := Game{
		Id:        gameID,
		Words:     []string{},
		BagTiles:  gameTiles[nPlayTiles:],
		UsedTiles: []Tile{},
		PlayTiles: gameTiles[:nPlayTiles],
		Points:    0,
	}
	gamesMutex.Lock()
	currentGames[gameID] = newGame
	gamesMutex.Unlock()
	newGame.PrintGame()
	response := GameResponse{
		GameId: newGame.Id,
		Tiles:  newGame.PlayTiles,
		Points: newGame.Points,
	}
	return c.JSON(http.StatusOK, response)
}

func guessWord(c echo.Context) error {

	g := new(GuessRequest)
	if err := c.Bind(g); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	gameId := g.Game
	guess := g.Guess

	if _, exists := wordMap[guess]; !exists {
		return c.JSON(http.StatusOK, BadGuessResponse{
			Valid: false,
		})
	}

	gamesMutex.Lock()
	game := currentGames[gameId] // add verification for situations where the game is not found
	guessLetters := []rune(strings.ToUpper(guess))
	usedTiles := make([]Tile, 0)
	unusedTiles := game.PlayTiles

	for _, letter := range guessLetters {
		for i, tile := range unusedTiles {
			if rune(tile.Letter) == letter {
				usedTiles = append(usedTiles, tile)
				unusedTiles = append(unusedTiles[:i], unusedTiles[i+1:]...)
				break
			}
		}
	}

	if len(guessLetters) != len(usedTiles) {
		return c.JSON(http.StatusOK, BadGuessResponse{
			Valid: false,
		})
	}

	guessPoints := 0
	for _, tile := range usedTiles {
		guessPoints += tile.Points
	}
	game.Points += guessPoints
	game.UsedTiles = append(game.UsedTiles, usedTiles...)
	game.Words = append(game.Words, guess)
	if len(game.BagTiles) > (nPlayTiles - len(usedTiles)) {
		game.PlayTiles = append(unusedTiles, game.BagTiles[:len(usedTiles)]...)
		game.BagTiles = game.BagTiles[len(usedTiles):]
	} else {
		game.PlayTiles = append(unusedTiles, game.BagTiles...)
		game.BagTiles = []Tile{}
		//NO MORE TILES!!
	}

	// Add points to game total
	currentGames[gameId] = game //check if needed
	gamesMutex.Unlock()

	//check if the guess is a word and if so:
	//calculate points and add to game
	//shuffle the bag and get new tiles
	//send the new set of 7 tiles

	game.PrintGame()

	response := GoodGuessResponse{
		GameId: game.Id,
		Tiles:  game.PlayTiles,
		Points: game.Points,
		Valid:  true,
	}
	return c.JSON(http.StatusOK, response)
}

func getNewTiles(c echo.Context) error {
	gameId := c.QueryParam("name")

	//TODO: extract repeated logic
	gamesMutex.Lock()
	game := currentGames[gameId] // add verification for situations where the game is not found
	currentPlayTiles := game.PlayTiles
	game.PlayTiles = game.BagTiles[:nPlayTiles] //<- crashing here
	game.BagTiles = append(game.BagTiles[nPlayTiles:], currentPlayTiles...)
	currentGames[gameId] = game //check if needed
	gamesMutex.Unlock()

	game.PrintGame()

	response := GoodGuessResponse{
		GameId: game.Id,
		Tiles:  game.PlayTiles,
		Points: game.Points,
		Valid:  true,
	}
	return c.JSON(http.StatusOK, response)
}

func endGame(c echo.Context) error {
	gameId := c.QueryParam("name")
	gamesMutex.Lock()
	game := currentGames[gameId]
	delete(currentGames, gameId)
	gamesMutex.Unlock()
	response := GameResponse{
		GameId: game.Id,
		Tiles:  []Tile{{'T', 1}, {'I', 1}, {'M', 3}, {'E', 1}, {'S', 1}, {'U', 1}, {'P', 3}},
		Points: game.Points,
	}
	return c.JSON(http.StatusOK, response)
}

// ShuffleTiles shuffles a slice of tiles in place using the Fisher-Yates algorithm.
func shuffleTiles(tiles *[]Tile) {
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	r := rand.New(source)
	n := len(*tiles)
	for i := n - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		(*tiles)[i], (*tiles)[j] = (*tiles)[j], (*tiles)[i]
	}
}
