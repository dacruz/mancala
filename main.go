package main

import (
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	ENV_REDIS_ADDRESS= "REDIS_ADDRESS"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

    redisAddr, redisAddrExists := os.LookupEnv(ENV_REDIS_ADDRESS)
    if !redisAddrExists {
        redisAddr = "localhost:6379"
    }

	p := &redis.Pool{
        MaxIdle:     10,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", redisAddr)
            if err != nil {
                return nil, err
            }
            return c, err
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
    }

	r := newMatchRepo(p)
    
	d := newDealer(r)

	if err := startServer(d); err != nil {
		log.Fatal(err)
	}
}
