package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

const (
	playerCookieConst = "player_id"
)

type ErrorMessage struct {
	Message string `json:"error"`
}

type MatchResponse struct {
	Id     string  `json:"match"`
	Board   [][]int `json:"board"`
	MyTurn bool    `json:"my_turn"`
}

type Handler struct {
	dealer Dealer	
}

func startServer(d Dealer) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	h := Handler{dealer: d}
	

	router := httprouter.New()
	router.GET("/", h.joinMatch)
	router.GET("/:matchId", h.getMatch)
	router.PUT("/:matchId/:pit", h.move)

	return http.ListenAndServe(":8080", router)
}

func (h Handler) joinMatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer handle5xx(w)
	defer setContectType(w)

	match, playerId := h.dealer.JoinMatch()

	response := MatchResponse{Id: match.Id, Board: match.Board, MyTurn: false}
	bs, _ := json.Marshal(response)

	cookie := &http.Cookie{
		Name:  playerCookieConst,
		Value: playerId,
	}
	http.SetCookie(w, cookie)

	w.Write(bs)
}

func (h Handler) getMatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer handle5xx(w)
	defer setContectType(w)

	matchIdParam := ps.ByName("matchId")

	playerCookie, err := r.Cookie(playerCookieConst)
	if err != nil {
		log.Print("ERROR - player cookie missing")
		writeErrorResponse("not your match", http.StatusUnauthorized, w)
		return
	}
	http.SetCookie(w, playerCookie)

	match, err := h.dealer.GetMatch(matchIdParam, playerCookie.Value)
	if err != nil {
		log.Printf("ERROR - match %v not found: %v", matchIdParam, err)
		msg := fmt.Sprintf("match %v not found", matchIdParam)
		writeErrorResponse(msg, http.StatusNotFound, w)
		return
	}

	myTurn := h.dealer.PlayerTurn(*match, playerCookie.Value)

	if playerCookie.Value == match.P2 {
		tmp := match.Board[0]
		match.Board[0] = match.Board[1]
		match.Board[1] = tmp
	}

	response := MatchResponse{Id: match.Id, Board: match.Board, MyTurn: myTurn}
	bs, _ := json.Marshal(response)
	w.Write(bs)
}

func (h Handler) move(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer handle5xx(w)
	defer setContectType(w)

	matchIdParam := ps.ByName("matchId")
	pit, _ := strconv.Atoi(ps.ByName("pit"))

	playerCookie, err := r.Cookie(playerCookieConst)
	if err != nil {
		log.Print("ERROR - player cookie missing")
		writeErrorResponse("not your match", http.StatusUnauthorized, w)
		return
	}
	http.SetCookie(w, playerCookie)

	m, err := h.dealer.GetMatch(matchIdParam, playerCookie.Value)
	if err != nil {
		log.Printf("ERROR - match %v not found: %v", matchIdParam, err)
		msg := fmt.Sprintf("match %v not found", matchIdParam)
		writeErrorResponse(msg, http.StatusNotFound, w)
		return
	}

	validMove, err := h.dealer.MakeMove(pit, *m, playerCookie.Value)
	if err != nil {
		log.Printf("ERROR - invalid move: %v", err)
		writeErrorResponse("invalid move", http.StatusBadRequest, w)
		return
	}

	if validMove {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setContectType(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func writeErrorResponse(msg string, statusCode int, w http.ResponseWriter) {
	errMsg := ErrorMessage{Message: msg}
	errorBs, _ := json.Marshal(errMsg)
	w.WriteHeader(statusCode)
	w.Write(errorBs)
}

func handle5xx(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Println("ERROR - InternalServerError: ", r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
