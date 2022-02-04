package main

import (
	"errors"

	"github.com/google/uuid"
)

const (
	numStones  int = 6
	boardSize  int = 6
	numWorkers int = 100
	bufferSize int = 3
)

type Move struct {
	pit   int
	match Match
}

type MancalaBoard [][]int

type Match struct {
	Id   string
	Board MancalaBoard
	P1   string
	P2   string
	Turn string
}

type Dealer interface {
	JoinMatch() (*Match, string)
	GetMatch(string, string) (*Match, error)
	PlayerTurn(Match, string) bool
	MakeMove(int, Match, string) (bool, error)
}

type MancalaDealer struct {
	repo     MatchRepo
	moveInCh chan Move
}

func newDealer(r MatchRepo) Dealer {
	moveInCh := make(chan Move, boardSize)
	moveOutCh := make(chan Move, boardSize)
	d := MancalaDealer{repo: r, moveInCh: moveInCh}

	for i := 0; i < numWorkers; i++ {
		go handleMove(&d, moveOutCh)
		go handleMoveCompleted(&d, moveOutCh)
	}

	return &d
}

func (d *MancalaDealer) JoinMatch() (*Match, string) {

	m, err := d.repo.GetWaitingMatch()
	if err == nil {
		m.P2 = uuid.NewString()
		m.Turn = m.P1
		d.repo.Save(m)
		return m, m.P2
	}

	newMatch := Match{Id: uuid.NewString(), P1: uuid.NewString(), Board: newBoard()}
	d.repo.AddWaitingMatch(&newMatch)

	return &newMatch, newMatch.P1
}

func (d *MancalaDealer) GetMatch(matchId string, playerId string) (*Match, error) {
	match, err := d.repo.Get(matchId)
	if err != nil {
		return nil, errors.New("unnable to get match")
	}
	return match, nil
}

func (d *MancalaDealer) PlayerTurn(match Match, playerId string) bool {
	p1IsDone := true
	for i := 0; i < boardSize; i++ {
		if match.Board[0][i] != 0 {
			p1IsDone = false
			break
		}
	}

	p2IsDone := true
	for i := 0; i < boardSize; i++ {
		if match.Board[1][i] != 0 {
			p2IsDone = false
			break
		}
	}
	

	return !p1IsDone && !p2IsDone && match.Turn == playerId
}

func (d *MancalaDealer) MakeMove(pit int, match Match, playerId string) (bool, error) {
	if err := d.repo.Lock(match.Id); err != nil {
		return false, nil
	}

	if 0 > pit || pit >= boardSize {
		return false, errors.New("invalid pit number")
	}

	if match.Turn != playerId {
		return false, nil

	}

	d.moveInCh <- Move{pit, match}

	return true, nil

}

func newBoard() [][]int {
	b := make([][]int, 2)
	for i := 0; i < boardSize; i++ {
		b[0] = append(b[0], numStones)
		b[1] = append(b[1], numStones)
	}

	b[0] = append(b[0], 0)
	b[1] = append(b[1], 0)

	return b
}

func handleMove(d *MancalaDealer, out chan Move) {
	for m := range d.moveInCh {
		executeMove(m, d.moveInCh, out)
	}
}

func handleMoveCompleted(d *MancalaDealer, ch chan Move) {
	for m := range ch {
		if m.match.Turn == m.match.P1 {
			m.match.Turn = m.match.P2
		} else {
			m.match.Turn = m.match.P1
		}

		d.repo.Save(&m.match)
	}
}

func executeMove(move Move, moveCh chan Move, completedCh chan Move) {
	var playerBoard int
	var opponentBoard int

	if move.match.Turn == move.match.P1 {
		playerBoard = 0
		opponentBoard = 1
	} else {
		playerBoard = 1
		opponentBoard = 0
	}

	linearBoard := append(move.match.Board[playerBoard], move.match.Board[opponentBoard]...)

	stones := linearBoard[move.pit]
	linearBoard[move.pit] = 0

	firstPit := move.pit + 1
	lastPit := len(linearBoard)-1
	for {
		for i := firstPit; i < lastPit; i++ {
			linearBoard[i] += 1
			stones -= 1
			if stones == 0 {
				move.match.Board[playerBoard] = linearBoard[:boardSize+1]
				move.match.Board[opponentBoard] = linearBoard[boardSize+1:]

				if endsOnMySide(i){
					if wasEmpty(linearBoard[i]) {
						collectStones(i,linearBoard)
						completedCh <- move
					} else {
						moveCh <- Move{pit: i, match: move.match}
					}

				} else {
					completedCh <- move
				}
				return
			}
		}
		firstPit = 0
	}
}

func collectStones(i int, linearBoard []int) {
	linearBoard[i] = 0
	opponentStone := linearBoard[i+boardSize+1]
	linearBoard[i+boardSize+1] = 0
	linearBoard[boardSize] += opponentStone + 1
}

func wasEmpty(i int) bool {
	return i == 1
}

func endsOnMySide(i int) bool {
	return i < boardSize 
}
