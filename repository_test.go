package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/rafaeljusto/redigomock"
)

func TestAddWaitingMatchSavesMatchAndWaitingId(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	m := Match{Id: uuid.NewString()}
	matchValue, _ := json.Marshal(m)
	matchKey := fmt.Sprintf("match:%v", m.Id)

	conn.Command("MULTI").Expect("OK")
	conn.Command("SET", matchKey, matchValue).Expect("OK")
	conn.Command("SADD", "waiting_match", m.Id).Expect("OK")
	conn.Command("EXEC").Expect("OK")

	repo.AddWaitingMatch(&m)

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestGet(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	m := Match{Id: uuid.NewString()}

	matchValue, _ := json.Marshal(m)
	matchKey := fmt.Sprintf("match:%v", m.Id)
	conn.Command("GET", matchKey).Expect(matchValue)

	repo.Get(m.Id)

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestGetFail(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	mId := uuid.NewString()

	matchKey := fmt.Sprintf("match:%v", mId)
	conn.Command("GET", matchKey).ExpectError(errors.New("erro"))

	_, err := repo.Get(mId)
	if err == nil {
		t.Fatalf("error expected")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestGetWaitingMatch(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	m := Match{Id: uuid.NewString()}
	conn.Command("SPOP", "waiting_match").Expect(m.Id)

	matchValue, _ := json.Marshal(m)
	matchKey := fmt.Sprintf("match:%v", m.Id)
	conn.Command("GET", matchKey).Expect(matchValue)

	repo.GetWaitingMatch()

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestGetWaitingMatchFail(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	conn.Command("SPOP", "waiting_match").ExpectError(errors.New("error"))

	_, err := repo.GetWaitingMatch()
	if err == nil {
		t.Fatalf("error expected")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestSaveMatch(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	m := Match{Id: uuid.NewString()}

	matchValue, _ := json.Marshal(m)
	matchKey := fmt.Sprintf("match:%v", m.Id)
	conn.Command("SET", matchKey, matchValue).Expect("Ok!")

	repo.Save(&m)

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestLockMatch(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	mId := uuid.NewString()

	conn.Command("SET", mId, "locked", "EX", 1, "NX").Expect("Ok!")

	repo.Lock(mId)

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}

func TestFailLockMatch(t *testing.T) {
	conn := redigomock.NewConn()
	repo := newMatchRepo(&redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	})

	mId := uuid.NewString()

	conn.Command("SET", mId, "locked", "EX", 1, "NX").ExpectError(errors.New("error"))

	err := repo.Lock(mId)
	if err == nil {
		t.Fatalf("error expected")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}

}
