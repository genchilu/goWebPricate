package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	//	"reflect"
	"strconv"
	"time"
)

type Session struct {
	Sid          string                 // unique session id
	TimeAccessed time.Time              // last access time
	Value        map[string]interface{} // session value stored inside
}

func main() {
	c, err := redis.Dial("tcp", "192.168.99.100:6379")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()
	v := make(map[string]interface{}, 0)
	v["user"] = "g7"
	v["count"] = 1
	ses := Session{Sid: "123", TimeAccessed: time.Now(), Value: v}
	sesBytes, err := json.Marshal(ses)
	if err != nil {
		fmt.Println("err at marshal")
		fmt.Println(err)
	}
	sesJson := string(sesBytes)
	c.Do("SET", "123", sesJson)
	value, err := redis.String(c.Do("GET", "123"))
	ses1 := Session{}
	json.Unmarshal([]byte(value), &ses1)
	fmt.Println(ses)
	fmt.Println(ses1)
	tmp := ses1.Value["count"]
	//count := reflect.ValueOf(ses1.Value["count"]).String()
	countStr := fmt.Sprint(tmp)
	count, err := strconv.Atoi(countStr)
	fmt.Println(count)
}
