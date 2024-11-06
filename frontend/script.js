let gameId;
let tilesContainer;
let guessContainer;
let pointsContainer;
let timerContainer;
let titleContainer;
let guess = [];
let tiles = [];
let points = 0;
let timeLeft = 60;
// let availableTiles = [];

document.addEventListener('DOMContentLoaded', () => {

    titleContainer = document.getElementById('title')
    tilesContainer = document.getElementById('tiles-row')
    guessContainer = document.getElementById('board-row')
    pointsContainer = document.getElementById('score')
    timerContainer = document.getElementById('timer')

    // Fetch initial tiles
    startGame();
});

document.addEventListener('keydown', (event) => {
    if (event.key === 'Enter') {
        if (guess.length > 0) {
            submitGuess()
        }
    } else if (event.key === 'Backspace') {
        if (guess.length > 0) {
            const lastTileId = guess[guess.length - 1].id
            guess = guess.slice(0, -1)
            for (let i = tiles.length - 1; i >= 0; i--) {
                if (tiles[i].picked && tiles[i].id === lastTileId) {
                    tiles[i].picked = false
                    break
                }
            }
            renderGame()
        }
    } else if (event.key === ' ') {
        getNewTiles()
    } else {
        letterPressed = event.key.toUpperCase()
        let letterGotPicked = false
        for (let i = 0; i < tiles.length; i++) {
            if (String.fromCodePoint(tiles[i].letter) === letterPressed && !tiles[i].picked) {
                tiles[i].picked = true
                letterGotPicked = true
                guess.push(tiles[i])
                break
            }
        }
        if (letterGotPicked) 
            renderGame()
    }
});

window.addEventListener('beforeunload', (event) => {
    endGame()
});

// Fetch tiles from the backend
const startGame = () => {
    fetch('http://localhost:8080/start')
        .then(response => response.json())
        .then(data => {
            gameId = data['gameId']
            tiles = data['tiles'].map((tile, index) => ({id: index, points: tile.Points, letter: tile.Letter, picked: false}));
            points = data['points']
            renderGame()
            startTimer()
        })
        .catch(error => console.error('Error fetching tiles:', error))
}

const getNewTiles = () => {
    fetch(`http://localhost:8080/tiles?name=${gameId}`)
        .then(response => response.json())
        .then(data => {
            guess = []
            tiles = data['tiles'].map((tile, index) => ({id: index, points: tile.Points, letter: tile.Letter, picked: false}));
            renderGame()
            changeTimeLeft(-5)
        })
        .catch(error => console.error('Error fetching tiles:', error))
}

const renderGame = () => {
    console.log('TILES -> ', tiles)
    renderGuess()
    renderTiles()
    pointsContainer.innerText = `🧮 ${points} pts`
}
const resetTitle = () => {
    titleContainer.innerHTML = '<h1>SPEEDRABBLE</h1>'
    titleContainer.querySelector('h1').style.backgroundColor = 'red';

}

// Render tiles
const renderTiles = () => {
    tilesContainer.innerHTML = ''
    tiles.forEach(tile => {
        const tileDiv = document.createElement('div')
        if (tile.picked) {
            tileDiv.classList.add('empty-square')
        } else {
            tileDiv.classList.add('tile')
            const letterSpan = document.createElement('span')
            letterSpan.classList.add('letter')
            letterSpan.textContent = String.fromCodePoint(tile.letter) 
            const pointsSpan = document.createElement('span')
            pointsSpan.classList.add('points') 
            pointsSpan.textContent = tile.points
            tileDiv.appendChild(letterSpan)
            tileDiv.appendChild(pointsSpan)
        }
        tilesContainer.appendChild(tileDiv)
    });
}

const renderGuess = () => {
    guessContainer.innerHTML = ''
    guess.forEach(tile => {
        const tileDiv = document.createElement('div')
        tileDiv.classList.add('tile')
        const letterSpan = document.createElement('span')
        letterSpan.classList.add('letter')
        letterSpan.textContent = String.fromCodePoint(tile.letter) 
        const pointsSpan = document.createElement('span')
        pointsSpan.classList.add('points') 
        pointsSpan.textContent = tile.points
        tileDiv.appendChild(letterSpan)
        tileDiv.appendChild(pointsSpan)
        guessContainer.appendChild(tileDiv)
    });
    for (let i = 0; i < 7 - guess.length; i++) {
        const squareDiv = document.createElement('div')
        squareDiv.classList.add('square')
        guessContainer.appendChild(squareDiv)
    }
}

// Submit guess to the backend
const submitGuess = () => {
    if (guess.length === 0 && timeLeft > 0) {
        return
    }
    
    guessToSend = guess.map(tile => String.fromCodePoint(tile.letter)).join('').toLowerCase()
    console.log(guessToSend)
    //send the gameId to identify the game too
    fetch('http://localhost:8080/guess', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ game: gameId, guess: guessToSend })
    })
    .then(response => response.json())
    .then(data => {
        if (!data['valid']) {
            titleContainer.innerHTML = '<h1>NICE TRY...</h1>'
            titleContainer.querySelector('h1').style.backgroundColor = '#35654d';
        } else {
            titleContainer.innerHTML = '<h1>GOOD ONE!</h1>'
            titleContainer.querySelector('h1').style.backgroundColor = '#35654d';
            guess = []
            tiles = data['tiles'].map((tile, index) => ({id: index, points: tile.Points, letter: tile.Letter, picked: false}));
            points = data['points']
            renderGame()
            changeTimeLeft(10)
        }
        setTimeout(() => resetTitle(), 1000)
    })
    .catch(error => console.error('Error submitting guess:', error))
}

const startTimer = () => {
    updateTimerDisplay();
    timerInterval = setInterval(() => {
        timeLeft--;
        updateTimerDisplay();
        
        if (timeLeft <= 0) {
            clearInterval(timerInterval);
            endGame();
        }
    }, 1000);
}

const changeTimeLeft = (seconds) => {
    timeLeft += seconds
    updateTimerDisplay()
}

const updateTimerDisplay = () => {
    timerContainer.innerText = `⏱️ ${timeLeft}s`
}

const endGame = () => {
    // Disable further gameplay
    fetch(`http://localhost:8080/end?name=${gameId}`)
        .then(response => response.json())
        .then(data => {
            guess = []
            tiles = data['tiles'].map((tile, index) => ({id: index, points: tile.Points, letter: tile.Letter, picked: false}));
            points = data['points']
            renderGame()
            timerContainer.innerText = 'Game over!'
        })
        .catch(error => console.error('Error fetching tiles:', error))
}