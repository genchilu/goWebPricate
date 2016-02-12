package redissession

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/genchilu/goWebPricate/session"
	"sync"
)

var MaxLifeTime int64 = 10
var Pder = &Provider{list: list.New()}
var RedisCon redis.Conn

type SessionRedis struct {
	Sid   string                 // unique session id
	Value map[string]interface{} // session value stored inside
}

func (sr *SessionRedis) Set(key string, value interface{}) error {
	sr.Value[key] = value
	Pder.SessionUpdate(sr)
	return nil
}

func (sr *SessionRedis) Get(key string) interface{} {
	Pder.SessionUpdate(sr)
	if v, ok := sr.Value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (sr *SessionRedis) Delete(key string) error {
	delete(sr.Value, key)
	Pder.SessionUpdate(sr)
	return nil
}

func (sr *SessionRedis) SessionID() string {
	return sr.Sid
}

type Provider struct {
	lock     sync.Mutex               // lock
	Sessions map[string]*list.Element // save in memory
	list     *list.List               // gc
}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	v := make(map[string]interface{}, 0)
	newsess := &SessionRedis{Sid: sid, Value: v}
	pder.SessionUpdate(newsess)
	return newsess, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	value, err := redis.String(RedisCon.Do("GET", sid))
	if err != nil {
		sess, err := pder.SessionInit(sid)
		return sess, err
	}
	sr := SessionRedis{}
	json.Unmarshal([]byte(value), &sr)
	return &sr, nil
}

func (pder *Provider) SessionDestroy(sid string) error {
	RedisCon.Do("DEL", sid)
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	/*use EXPIRE cmd at redis for session gc*/
}

func (pder *Provider) SessionUpdate(session *SessionRedis) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	sesBytes, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sesJson := string(sesBytes)
	status, err := RedisCon.Do("SET", session.Sid, sesJson)
	if err != nil {
		fmt.Println(status)
		panic(err)
	}
	fmt.Printf("reset sid expire: %s\n", session.Sid)
	status, err = RedisCon.Do("EXPIRE", session.Sid, MaxLifeTime)
	if err != nil {
		fmt.Println(status)
		panic(err)
	}
	return err
}
