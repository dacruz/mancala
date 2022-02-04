package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"

	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type StubDealer struct{}

func init() {
	go startServer(&StubDealer{})
	time.Sleep(0)
}

func TestJoinMatch(t *testing.T) {
	res := execute2xxRequest("GET", "http://localhost:8080", t)

	bs, _ := ioutil.ReadAll(res.Body)

	match := MatchResponse{}
	err := json.Unmarshal(bs, &match)
	if err != nil {
		t.Fatalf("invalid response: %v", string(bs))
	}

	p1Cookie := getCookieByName(playerCookieConst, res.Cookies())
	_, err = uuid.Parse(p1Cookie.Value)
	if err != nil {
		t.Fatalf("player cookie not set. expected valid UUID but got %v", p1Cookie)
	}
}

func TestGetMatch(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res := execute2xxRequest("GET", url, t, &cookie)

	bs, _ := ioutil.ReadAll(res.Body)

	matchResponse := MatchResponse{}
	json.Unmarshal(bs, &matchResponse)

	if matchResponse.Id != testMatch.Id {
		t.Fatalf("wrong match. expected %v but got %v", testMatch.Id, matchResponse.Id)
	}

}

func TestGetMatchAsP2RotatesBoard(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P2}
	res := execute2xxRequest("GET", url, t, &cookie)

	bs, _ := ioutil.ReadAll(res.Body)

	matchResponse := MatchResponse{}
	json.Unmarshal(bs, &matchResponse)

	if matchResponse.Board[0][0] != 1 {
		t.Fatalf("board was not rotated. expected %v but got %v", testMatch.Board, matchResponse.Board)
	}
	if matchResponse.Board[1][0] != 0 {
		t.Fatalf("board was not rotated. expected %v but got %v", testMatch.Board, matchResponse.Board)
	}
}

func TestGetMatchMissingCookie(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	res := execute4xxRequest("GET", url, t)

	if res.StatusCode != 401 {
		t.Fatalf("expected 401, but got status code %v", res.StatusCode)
	}

	bs, _ := ioutil.ReadAll(res.Body)
	err := ErrorMessage{}
	json.Unmarshal(bs, &err)

	if err.Message != "not your match" {
		t.Fatalf("expected  \"not your match\" messsage but got %v", err.Message)
	}
}

func TestGetNotMyMatch(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	cookie := http.Cookie{Name: playerCookieConst, Value: uuid.NewString()}
	res := execute4xxRequest("GET", url, t, &cookie)

	if res.StatusCode != 404 {
		t.Fatalf("expected 404, but got status code %v", res.StatusCode)
	}

	bs, _ := ioutil.ReadAll(res.Body)
	err := ErrorMessage{}
	json.Unmarshal(bs, &err)

	expectedMsg := fmt.Sprintf("match %v not found", testMatch.Id)
	if err.Message != expectedMsg {
		t.Fatalf("wrong error message: %v", err.Message)
	}

}

func TestGetMatchOnMyTurn(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res := execute2xxRequest("GET", url, t, &cookie)

	bs, _ := ioutil.ReadAll(res.Body)

	matchResponse := MatchResponse{}
	json.Unmarshal(bs, &matchResponse)

	if !matchResponse.MyTurn {
		t.Fatal("it should by my turn")
	}

}

func TestGetMatchKeepsCookie(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", testMatch.Id)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res := execute2xxRequest("GET", url, t, &cookie)

	resCookie := getCookieByName(playerCookieConst, res.Cookies())

	if resCookie.Value != cookie.Value {
		t.Fatalf("player cookie is different. expected: %v but got %v", cookie.Value, resCookie.Value)
	}

}

func TestGetMatchWithUnkownId(t *testing.T) {
	id := uuid.New()
	url := fmt.Sprintf("http://localhost:8080/%v", id)
	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}

	res := execute4xxRequest("GET", url, t, &cookie)

	if res.StatusCode != 404 {
		t.Fatalf("expected 404, but got status code %v", res.StatusCode)
	}

	bs, _ := ioutil.ReadAll(res.Body)
	err := ErrorMessage{}
	json.Unmarshal(bs, &err)

	expectedMsg := fmt.Sprintf("match %v not found", id)
	if err.Message != expectedMsg {
		t.Fatalf("wrong error message: %v", err.Message)
	}

}

func Test5xxHandler(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v", panicGenerator)
	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res, err := doRequest("GET", url, []*http.Cookie{&cookie})
	if err != nil {
		t.Fatal("failed to execute http get")
	}

	if res.StatusCode != 500 {
		t.Fatalf("expected 500, but got status code %v", res.StatusCode)
	}
}

func TestMakeMove(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", testMatch.Id, 1)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}

	res := execute2xxRequest("PUT", url, t, &cookie)

	if res.StatusCode != 202 {
		t.Fatalf("expected 202, but got status code %v", res.StatusCode)
	}
}

func TestMakeMoveNotMyTurn(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", testMatch.Id, 1)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P2}

	res := execute2xxRequest("PUT", url, t, &cookie)

	if res.StatusCode != 204 {
		t.Fatalf("expected 204, but got status code %v", res.StatusCode)
	}
}

func TestMakeMoveNoPlayerCookie(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", testMatch.Id, 1)

	res := execute4xxRequest("PUT", url, t)

	if res.StatusCode != 401 {
		t.Fatalf("expected 401, but got status code %v", res.StatusCode)
	}
}

func TestMakeMoveNoMatch(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", uuid.New(), 1)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P2}
	res := execute4xxRequest("PUT", url, t, &cookie)

	if res.StatusCode != 404 {
		t.Fatalf("expected 404, but got status code %v", res.StatusCode)
	}
}

func TestMakeInvalidMove(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", testMatch.Id, -1)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res := execute4xxRequest("PUT", url, t, &cookie)

	if res.StatusCode != 400 {
		t.Fatalf("expected 400, but got status code %v", res.StatusCode)
	}
}

func TestMakeMoveKeepsCookie(t *testing.T) {
	url := fmt.Sprintf("http://localhost:8080/%v/%v", testMatch.Id, 1)

	cookie := http.Cookie{Name: playerCookieConst, Value: testMatch.P1}
	res := execute2xxRequest("PUT", url, t, &cookie)

	resCookie := getCookieByName(playerCookieConst, res.Cookies())

	if resCookie.Value != cookie.Value {
		t.Fatalf("player cookie is different. expected: %v but got %v", cookie.Value, resCookie.Value)
	}

}

func getCookieByName(name string, cl []*http.Cookie) *http.Cookie {
	for _, c := range cl {
		if c.Name == name {
			return c
		}
	}

	return nil
}

func execute2xxRequest(method string, url string, t *testing.T, cookies ...*http.Cookie) *http.Response {
	res, err := doRequest(method, url, cookies)
	if err != nil {
		t.Fatal("failed to execute http get")
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("expected 2xx, but got status code %v", res.StatusCode)
	}

	return res
}

func execute4xxRequest(method string, url string, t *testing.T, cookies ...*http.Cookie) *http.Response {
	res, err := doRequest(method, url, cookies)
	if err != nil {
		t.Fatal("failed to execute http get")
	}

	if res.StatusCode < 400 || res.StatusCode > 499 {
		t.Fatalf("expected 4xx, but got status code %v", res.StatusCode)
	}

	return res
}

func doRequest(method string, url string, cookies []*http.Cookie) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.New("failed to execute http request")
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	return client.Do(req)
}

var panicGenerator = uuid.NewString()
var testMatch = Match{Id: uuid.NewString(), P1: uuid.NewString(), P2: uuid.NewString(), Board: [][]int{{0,0},{1,1}}}

func (s *StubDealer) JoinMatch() (*Match, string) {
	return &testMatch, uuid.NewString()
}

func (s *StubDealer) GetMatch(matchId string, playerId string) (*Match, error) {
	if testMatch.Id == matchId {
		if testMatch.P1 == playerId || testMatch.P2 == playerId {
			return &testMatch, nil
		}
	}

	if panicGenerator == matchId {
		panic("panic!")
	}

	return &Match{}, errors.New("not found")

}

func (s *StubDealer) PlayerTurn(match Match, playerId string) bool {
	return true
}

func (d *StubDealer) MakeMove(pit int, match Match, playerId string) (bool, error) {
	if pit < 0 {
		return false, errors.New("")
	}
	return match.P1 == playerId, nil
}
