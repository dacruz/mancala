package main

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type StubRepo struct {
	match        *Match
	waitingMatch *Match
}

func TestJoinNewMatch(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	match, p1 := md.JoinMatch()

	if match.P1 != p1 {
		t.Fatalf("Player1 does not match: %v and %v", match.P1, p1)
	}

}

func TestJoinExistingMatch(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	originalMatch, _ := md.JoinMatch()
	existingMatch, _ := md.JoinMatch()

	if originalMatch.Id != existingMatch.Id {
		t.Fatalf("A new match shoudl not be created if one is already available:\n original: %v \n new: %v", originalMatch.Id, existingMatch.Id)
	}

}

func TestJoinExistingMatchSetsP2(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	md.JoinMatch()
	existingMatch, p2 := md.JoinMatch()

	if existingMatch.P2 != p2 {
		t.Fatalf("Player2 does not match: %v and %v", existingMatch.P2, p2)
	}

}

func TestSetsTurnToP1WhenGameStarts(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	_, p1 := md.JoinMatch()
	match, p2 := md.JoinMatch()

	if !md.PlayerTurn(*match, p1) {
		t.Fatal("expected to be Player1 turn but it is not")
	}

	if md.PlayerTurn(*match, p2) {
		t.Fatal("expected to not be Player2 turn but it is")
	}

}

func TestSetsTurnToNoOneIfBoardAreEmpty(t *testing.T) {
	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	p1 := uuid.NewString()
	p2 := uuid.NewString()

	match := Match{P1: p1, P2: p2, Turn: p1, Board: MancalaBoard{{0,0,0,0,0,0,10}, {1,1,1,1,1,1,1}}}
	if md.PlayerTurn(match, match.P1) {
		t.Fatal("expected to not be Player1 turn but it is")
	}

	match = Match{P1: p1, P2: p2, Turn: p2, Board: MancalaBoard{{1,1,1,1,1,1,1}, {0,0,0,0,0,0,10}}}
	if md.PlayerTurn(match, match.P2) {
		t.Fatal("expected to not be Player2 turn but it is")
	}

}

func TestGetExistingMatch(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	match := Match{Id: uuid.NewString(), P1: uuid.NewString(), Board: newBoard()}
	stubRepo.Save(&match)

	existingMatch, _ := md.GetMatch(match.Id, match.P1)

	if existingMatch.Id != match.Id {
		t.Fatalf("failed to get correct match. expected %v but got %v", match.Id, existingMatch.Id)
	}

}

func TestGetNonExistingMatch(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	match := Match{Id: uuid.NewString(), P1: uuid.NewString(), Board: newBoard()}
	stubRepo.Save(&match)

	_, err := md.GetMatch(uuid.NewString(), uuid.NewString())
	if err == nil {
		t.Fatal("it should not get a non existing match")
	}

}

func TestMakeMoveOnMyTurn(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, Turn: p1Id, Board: newBoard()}

	validMove, _ := md.MakeMove(1, match, match.P1)

	if !validMove {
		t.Fatal("it should be a valid move")
	}

}

func TestMakeMoveNotOnMyTurn(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	match := Match{Id: uuid.NewString(), P1: uuid.NewString(), Turn: uuid.NewString(), Board: newBoard()}

	validMove, _ := md.MakeMove(1, match, match.P1)

	if validMove {
		t.Fatal("it should not be a valid move")
	}

}

func TestMakeMoveInvalidPit(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, Turn: p1Id, Board: newBoard()}

	_, err := md.MakeMove(-1, match, match.P1)
	if err == nil {
		t.Fatal("pit -1 should not be a valid move")
	}

	_, err = md.MakeMove(boardSize, match, match.P1)
	if err == nil {
		t.Fatal("pit greater than the numbers of pits should not be a valid move")
	}

	_, err = md.MakeMove(boardSize+1, match, match.P1)
	if err == nil {
		t.Fatal("pit greater than the numbers of pits should not be a valid move")
	}

}

func TestMakeMoveChangesBoard(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, Turn: p1Id, Board: newBoard()}

	md.MakeMove(0, match, match.P1)
	time.Sleep(1 * time.Millisecond)

	if match.Board[0][0] != 0 {
		t.Fatal("pit should be zero after the move")
	}

}

func TestMakeMoveOnOngoingMove(t *testing.T) {

	var stubRepo MatchRepo = &StubRepo{}
	md := newDealer(stubRepo)

	p1Id := uuid.NewString()
	match := Match{Id: lockedMatchId, P1: p1Id, Turn: p1Id, Board: newBoard()}

	validMove, _ := md.MakeMove(0, match, match.P1)
	if validMove {
		t.Fatal("no move should be made on a locked match")
	}

}

func TestMoveFinishesOnOpponentsPit(t *testing.T) {

	board := MancalaBoard{{0,0,0,0,0,3,0},{0,1,0,0,0,0,0}}

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, P2: uuid.NewString(), Turn: p1Id, Board: board}

	move := Move {
		pit: 5,
		match: match,
	}

	result := testMove(move)

	if result.match.Board[1][0] != 1 {
		t.Fatalf("expected board[1][0] == 1 but got:%v \n %v", result.match.Board[1][0], result.match.Board)
	}

	if result.match.Board[1][1] != 2 {
		t.Fatalf("expected board[1][1] == 2 but got:%v \n %v", result.match.Board[1][1], result.match.Board)
	}
	
	if result.match.Board[0][5] != 0 {
		t.Fatalf("expected board[0][5] == 0 but got:%v \n %v", result.match.Board[0][5], result.match.Board)
	}
}

func TestMoveSkipsOpponentsBigPit(t *testing.T) {

	board := MancalaBoard{{1,0,0,0,0,8,0},{1,0,0,0,0,0,0}}

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, P2: uuid.NewString(), Turn: p1Id, Board: board}

	move := Move {
		pit: 5,
		match: match,
	}

	result := testMove(move)

	if result.match.Board[1][6] != 0 {
		t.Fatalf("expected board[1][6] == 0 but got:%v \n %v", result.match.Board[1][6], result.match.Board)
	}
	
}

func TestMovePassTheBoarMoreThanOnce(t *testing.T) {

	board := MancalaBoard{{0,0,0,0,0,14,0},{0,0,0,0,0,1,0}}

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, P2: uuid.NewString(), Turn: p1Id, Board: board}

	move := Move {
		pit: 5,
		match: match,
	}

	result := testMove(move)

	if result.match.Board[0][0] != 1 {
		t.Fatalf("expected board[0][0] == 1 but got:%v \n %v", result.match.Board[0][0], result.match.Board)
	}
	
}

func TestMoveForP2(t *testing.T) {

	board := MancalaBoard{{0,0,0,0,0,1,0},{5,0,0,0,0,1,5}}

	p2Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: uuid.NewString(), P2: p2Id,  Turn: p2Id, Board: board}

	move := Move {
		pit: 5,
		match: match,
	}

	result := testMove(move)

	if result.match.Board[1][6] != 6 {
		t.Fatalf("expected board[1][6] == 6 but got:%v \n %v", result.match.Board[1][6], result.match.Board)
	}
	
}

func TestMoveFinishesOnPlayerBoarGetsAnotherTurn(t *testing.T) {

	board := MancalaBoard{{0,0,0,0,1,1,0},{0,0,0,0,0,1,0}}

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, P2: uuid.NewString(), Turn: p1Id, Board: board}

	move := Move {
		pit: 4,
		match: match,
	}

	result := testMove(move)
	
	if result.match.Board[0][4] != 0 {
		t.Fatalf("expected board[0][4] == 0 but got:%v \n %v", result.match.Board[0][4], result.match.Board)
	}

	if result.match.Board[0][5] != 0 {
		t.Fatalf("expected board[0][5] == 0 but got:%v \n %v", result.match.Board[0][5], result.match.Board)
	}

	if result.match.Board[0][6] != 1 {
		t.Fatalf("expected board[0][6] == 1 but got:%v \n %v", result.match.Board[0][6], result.match.Board)
	}

	if result.match.Board[1][0] != 1 {
		t.Fatalf("expected board[1][0] == 1 but got:%v \n %v", result.match.Board[1][0], result.match.Board)
	}

}

func TestMoveFinishesOnPlayerEmptyPitAndCaptureStones(t *testing.T) {

	board := MancalaBoard{{1,0,0,0,1,0,0},{0,0,0,0,1,10,0}}

	p1Id := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1Id, P2: uuid.NewString(), Turn: p1Id, Board: board}

	move := Move {
		pit: 4,
		match: match,
	}

	result := testMove(move)
	
	if result.match.Board[0][5] != 0 {
		t.Fatalf("expected board[0][5] == 0 but got:%v \n %v", result.match.Board[0][5], result.match.Board)
	}

	if result.match.Board[0][6] != 11 {
		t.Fatalf("expected board[0][6] == 11 but got:%v \n %v", result.match.Board[0][6], result.match.Board)
	}

	if result.match.Board[1][5] != 0 {
		t.Fatalf("expected board[1][5] == 0 but got:%v \n %v", result.match.Board[1][5], result.match.Board)
	}

}

func TestSavesMatchAfterMove(t *testing.T) {

	var stubRepo  = &StubRepo{}
	ch := make(chan Move, 10)
	d := MancalaDealer{repo: stubRepo}

	p1 := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1, P2: uuid.NewString(), Turn: p1}

	move := Move {
		pit: 4,
		match: match,
	}

	ch <- move
	close(ch)
	handleMoveCompleted(&d, ch)
	

	if stubRepo.match == nil {
		t.Fatalf("match was not saved")
	}
}

func TestChangeTurnsFromP1ToP2AfterMove(t *testing.T) {

	var stubRepo  = &StubRepo{}
	ch := make(chan Move, 10)
	d := MancalaDealer{repo: stubRepo}

	p1 := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: p1, P2: uuid.NewString(), Turn: p1}

	move := Move {
		pit: 4,
		match: match,
	}

	ch <- move
	close(ch)
	handleMoveCompleted(&d, ch)

	if stubRepo.match.Turn != stubRepo.match.P2 {
		t.Fatalf("it should now be p2 turn")
	}
}

func TestChangeTurnsFromP2ToP1AfterMove(t *testing.T) {

	var stubRepo  = &StubRepo{}
	ch := make(chan Move, 10)
	d := MancalaDealer{repo: stubRepo}

	p2 := uuid.NewString()
	match := Match{Id: uuid.NewString(), P1: uuid.NewString(), P2: p2, Turn: p2}

	move := Move {
		pit: 4,
		match: match,
	}

	ch <- move
	close(ch)
	handleMoveCompleted(&d, ch)

	if stubRepo.match.Turn != stubRepo.match.P1 {
		t.Fatalf("it should now be p1 turn")
	}
}


func testMove(m Move) (Move){
	in := make(chan Move, 1)
	out := make(chan Move, 1)

	var stubRepo  = &StubRepo{}

	d := MancalaDealer{repo: stubRepo, moveInCh: in}

	in <- m
	go handleMove(&d, out)

	return <- out 
}


func (r *StubRepo) Get(id string) (*Match, error) {
	if r.match.Id == id {
		return r.match, nil
	}

	return nil, errors.New("unnable to get match")
}

func (r *StubRepo) GetWaitingMatch() (*Match, error) {
	if r.waitingMatch != nil {
		m := r.waitingMatch
		r.waitingMatch = nil
		return m, nil
	}

	return nil, errors.New("unnable to get match")
}

func (r *StubRepo) AddWaitingMatch(match *Match) {
	r.waitingMatch = match
}

func (r *StubRepo) Save(match *Match) {
	r.match = match
}

var lockedMatchId = uuid.NewString()

func (r *StubRepo) Lock(id string) error {
	if lockedMatchId == id {
		return errors.New("already locked")
	}

	return nil
}
