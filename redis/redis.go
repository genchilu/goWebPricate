package redisSession

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/genchilu/goWebPricate/session"
	"sync"
	"time"
)

var pder = &Provider{list: list.New()}
var redisCon redis.Conn

type SessionRedis struct {
	Sid          string                 // unique session id
	TimeAccessed time.Time              // last access time
	Value        map[string]interface{} // session value stored inside
}

func (sr *SessionRedis) Set(key string, value interface{}) error {
	sr.Value[key] = value
	pder.SessionUpdate(sr)
	return nil
}

func (sr *SessionRedis) Get(key string) interface{} {
	pder.SessionUpdate(sr)
	if v, ok := sr.Value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

//func (sr *SessionRedis) Delete(key string) error {
//delete(sr.Value, key)
//pder.SessionUpdate(sr)
//	return nil
//}

func (sr *SessionRedis) Delete(key string) error {
	//delete(st.value, key)
	//pder.SessionUpdate(st.sid)
	return nil
}

func (sr *SessionRedis) SessionID() string {
	return sr.Sid
}

type Provider struct {
	lock     sync.Mutex               // lock
	sessions map[string]*list.Element // save in memory
	list     *list.List               // gc
}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	v := make(map[string]interface{}, 0)
	newsess := &SessionRedis{Sid: sid, Value: v}
	pder.SessionUpdate(newsess)
	return newsess, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	value, err := redis.String(redisCon.Do("GET", sid))
	if err != nil {
		sess, err := pder.SessionInit(sid)
		return sess, err
	}
	sr := SessionRedis{}
	json.Unmarshal([]byte(value), &sr)
	return &sr, nil
}

func (pder *Provider) SessionDestroy(sid string) error {
	redisCon.Do("DEL", sid)
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	/*
		pder.lock.Lock()
		defer pder.lock.Unlock()

		for {
			element := pder.list.Back()
			if element == nil {
				break
			}
			if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
				pder.list.Remove(element)
				delete(pder.sessions, element.Value.(*SessionStore).sid)
			} else {
				break
			}
		}
	*/
}

func (pder *Provider) SessionUpdate(session *SessionRedis) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	session.TimeAccessed = time.Now()
	sesBytes, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sesJson := string(sesBytes)
	status, err := redisCon.Do("SET", session.Sid, sesJson)
	if err != nil {
		fmt.Println(status)
		panic(err)
	}
	return err
}

func init() {
	pder.sessions = make(map[string]*list.Element, 0)
	con, err := redis.Dial("tcp", "192.168.99.100:6379")
	if err != nil {
		panic(err)
	}
	redisCon = con
	//defer redisCon.Close()
	session.Register("redis", pder)
	fmt.Println("finish init redis session")
}
