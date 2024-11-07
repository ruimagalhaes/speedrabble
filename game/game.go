package game

import (
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

var gamesMutex sync.Mutex

// Requests
type GuessRequest struct {
	Game  string `json:"game"`
	Guess string `json:"guess"`
}

type GameResponse struct {
	GameId string `json:"gameId"`
	Tiles  []Tile `json:"tiles"`
	Points int    `json:"points"`
}

// Responses
type GoodGuessResponse struct {
	GameId string `json:"gameId"`
	Tiles  []Tile `json:"tiles"`
	Points int    `json:"points"`
	Valid  bool   `json:"valid"`
}

type BadGuessResponse struct {
	Valid bool `json:"valid"`
}

// Models
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

type GameHandler struct {
	CurrentGames map[string]Game
	WordMap      map[string]int
	NPlayTiles   int
}

func (h *GameHandler) StartGame(c echo.Context) error {
	gameTiles := make([]Tile, len(scrabbleTiles))
	copy(gameTiles, scrabbleTiles)
	shuffleTiles(&gameTiles)
	gameID := generateGameID()
	newGame := Game{
		Id:        gameID,
		Words:     []string{},
		BagTiles:  gameTiles[h.NPlayTiles:],
		UsedTiles: []Tile{},
		PlayTiles: gameTiles[:h.NPlayTiles],
		Points:    0,
	}
	gamesMutex.Lock()
	h.CurrentGames[gameID] = newGame
	gamesMutex.Unlock()
	newGame.PrintGame()
	response := GameResponse{
		GameId: newGame.Id,
		Tiles:  newGame.PlayTiles,
		Points: newGame.Points,
	}
	return c.JSON(http.StatusOK, response)
}

func (h *GameHandler) GuessWord(c echo.Context) error {
	g := new(GuessRequest)
	if err := c.Bind(g); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	gameId := g.Game
	guess := g.Guess

	if _, exists := h.WordMap[guess]; !exists {
		return c.JSON(http.StatusOK, BadGuessResponse{
			Valid: false,
		})
	}

	gamesMutex.Lock()
	game := h.CurrentGames[gameId] // add verification for situations where the game is not found
	gLetters := []rune(strings.ToUpper(guess))
	gUsedTiles := make([]Tile, 0)
	gUnusedTiles := game.PlayTiles

	for _, letter := range gLetters {
		for i, tile := range gUnusedTiles {
			if rune(tile.Letter) == letter {
				gUsedTiles = append(gUsedTiles, tile)
				gUnusedTiles = append(gUnusedTiles[:i], gUnusedTiles[i+1:]...)
				break
			}
		}
	}

	if len(gLetters) != len(gUsedTiles) {
		return c.JSON(http.StatusOK, BadGuessResponse{
			Valid: false,
		})
	}

	guessPoints := 0
	for _, tile := range gUsedTiles {
		guessPoints += tile.Points
	}
	game.Points += guessPoints
	game.UsedTiles = append(game.UsedTiles, gUsedTiles...)
	game.Words = append(game.Words, guess)

	if len(game.BagTiles) > len(gUsedTiles) {
		game.PlayTiles = append(gUnusedTiles, game.BagTiles[:len(gUsedTiles)]...)
		game.BagTiles = game.BagTiles[len(gUsedTiles):]
	} else {
		game.PlayTiles = append(gUnusedTiles, game.BagTiles...)
		game.BagTiles = []Tile{}
	}

	h.CurrentGames[gameId] = game //check if needed
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

func (h *GameHandler) GetNewTiles(c echo.Context) error {
	gameId := c.QueryParam("name")

	gamesMutex.Lock()
	game := h.CurrentGames[gameId] // add verification for situations where the game is not found

	game.BagTiles = append(game.BagTiles, game.PlayTiles...)
	if len(game.BagTiles) > h.NPlayTiles {
		game.PlayTiles = game.BagTiles[:h.NPlayTiles]
		game.BagTiles = game.BagTiles[h.NPlayTiles:]
	} else {
		game.PlayTiles = game.BagTiles
		game.BagTiles = []Tile{}
	}

	h.CurrentGames[gameId] = game //check if needed
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

func (h *GameHandler) EndGame(c echo.Context) error {
	gameId := c.QueryParam("name")
	gamesMutex.Lock()
	game := h.CurrentGames[gameId]
	delete(h.CurrentGames, gameId)
	gamesMutex.Unlock()
	response := GameResponse{
		GameId: game.Id,
		Tiles:  timesupTiles,
		Points: game.Points,
	}
	return c.JSON(http.StatusOK, response)
}
