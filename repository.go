package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

type MatchRepo interface {
	Get(string) (*Match, error)
	Save(*Match)
	GetWaitingMatch() (*Match, error)
	AddWaitingMatch(*Match)
	Lock(id string) error
}

type RedisRepo struct {
	connPool *redis.Pool
}

func newMatchRepo(connPool *redis.Pool) MatchRepo {
	mr := RedisRepo{connPool: connPool}
	return &mr
}

func (r *RedisRepo) Get(id string) (*Match, error) {
	conn := r.connPool.Get()
	defer conn.Close()

	matchKey := fmt.Sprintf("match:%v", id)
	
	matchStr, err := redis.String(conn.Do("GET", matchKey))
	if err != nil {
		return nil, err
	}

	m := Match{}
	err = json.Unmarshal([]byte(matchStr), &m)

	return &m, err
}

func (r *RedisRepo) GetWaitingMatch() (*Match, error) {
	conn := r.connPool.Get()
	defer conn.Close()

	mId, err := redis.String(conn.Do("SPOP", "waiting_match"))
	if err != nil {
		log.Print("no waiting match - ", err)
		return nil, err
	}

	return r.Get(mId)
}

func (r *RedisRepo) AddWaitingMatch(m *Match) {
	matchKey := fmt.Sprintf("match:%v", m.Id)
	matchValue, err := json.Marshal(m)
	checkFatalError(err)

	conn := r.connPool.Get()
	defer conn.Close()

	err = conn.Send("MULTI")
	checkFatalError(err)

	err = conn.Send("SET", matchKey, matchValue)
	checkFatalError(err)

	err = conn.Send("SADD", "waiting_match", m.Id)
	checkFatalError(err)

	_, err = conn.Do("EXEC")
	checkFatalError(err)

}

func (r *RedisRepo) Save(m *Match) {
	matchKey := fmt.Sprintf("match:%v", m.Id)
	matchValue, err := json.Marshal(m)
	checkFatalError(err)

	conn := r.connPool.Get()
	defer conn.Close()

	_, err = conn.Do("SET", matchKey, matchValue)
	checkFatalError(err)
}

//I know this will not work well on a distributed Redis
// see: https://redis.io/topics/distlock
func (r *RedisRepo) Lock(id string) error {
	conn := r.connPool.Get()
	defer conn.Close()
	
	_, err := conn.Do("SET", id, "locked", "EX", 1, "NX")

	return err
}

//Better handling needed, but no time.
//http server on main.go will handle it and send a 500
func checkFatalError(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
