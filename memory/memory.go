package memory

import (
	"container/list"
	"github.com/genchilu/goWebPricate/session"
	"sync"
	"time"
)

var Pder = &Provider{list: list.New()}

type SessionStore struct {
	Sid          string                      // unique session id
	TimeAccessed time.Time                   // last access time
	Value        map[interface{}]interface{} // session value stored inside
}

func (st *SessionStore) Set(key string, value interface{}) error {
	st.Value[key] = value
	Pder.SessionUpdate(st.Sid)
	return nil
}

func (st *SessionStore) Get(key string) interface{} {
	Pder.SessionUpdate(st.Sid)
	if v, ok := st.Value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (st *SessionStore) Delete(key string) error {
	delete(st.Value, key)
	Pder.SessionUpdate(st.Sid)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.Sid
}

type Provider struct {
	lock     sync.Mutex               // lock
	Sessions map[string]*list.Element // save in memory
	list     *list.List               // gc
}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{Sid: sid, TimeAccessed: time.Now(), Value: v}
	element := pder.list.PushBack(newsess)
	pder.Sessions[sid] = element
	return newsess, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := pder.Sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := pder.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

func (pder *Provider) SessionDestroy(sid string) error {
	if element, ok := pder.Sessions[sid]; ok {
		delete(pder.Sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).TimeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			pder.list.Remove(element)
			delete(pder.Sessions, element.Value.(*SessionStore).Sid)
		} else {
			break
		}
	}
}

func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.Sessions[sid]; ok {
		element.Value.(*SessionStore).TimeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}
