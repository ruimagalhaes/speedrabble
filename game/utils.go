package game

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

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

func (g *Game) PrintGame() {
	fmt.Println("===================================")
	fmt.Printf("GAME %s:\n", g.Id)
	bagStr := "BAG TILES"
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
