package main

import (
	"container/list"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/genchilu/goWebPricate/memory"
	"github.com/genchilu/goWebPricate/redissession"
	"os"
	//"github.com/fvbock/endless"
	"github.com/genchilu/goWebPricate/session"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
)

var globalSessions *session.Manager
var hostname string

// Then, initialize the session manager
func init() {
	//parser arg
	var sessionType string
	var redisIpAndPort string
	var maxLifeTime int64
	flag.StringVar(&sessionType, "sessiontype", "memory", "session type (memory or redis)")
	flag.StringVar(&redisIpAndPort, "redisinfo", "redis:6379", "ip and prot of redis")
	flag.Int64Var(&maxLifeTime, "sessionlifetime", 10, "session life time in secend")
	flag.Parse()
	fmt.Printf("session type: %s\n", sessionType)
	//init memory session
	memory.Pder.Sessions = make(map[string]*list.Element, 0)
	session.Register("memory", memory.Pder)
	fmt.Println("finish init memory session")
	//init redis session
	redissession.MaxLifeTime = maxLifeTime
	redissession.Pder.Sessions = make(map[string]*list.Element, 0)
	var err error
	if sessionType == "redis" {
		redissession.RedisCon, err = redis.Dial("tcp", redisIpAndPort)
		if err != nil {
			panic(err)
		}
		session.Register("redis", redissession.Pder)
		fmt.Println("finish init redis session")
	}
	globalSessions, _ = session.NewManager(sessionType, "gosessionid", maxLifeTime)
	if sessionType == "memory" {
		go globalSessions.GC()
	}
	hostname, _ = os.Hostname()
	fmt.Println("finish init main")
}

func login(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		if sess.Get("username") != nil {
			http.Redirect(w, r, "/", 302)
		} else {
			fmt.Println("render login page")
			t, _ := template.ParseFiles("login.gtpl")
			w.Header().Set("Content-Type", "text/html")
			t.Execute(w, sess.Get("username"))
		}
	} else {
		sess.Set("username", r.FormValue("username"))
		sess.Set("count", 1)
		http.Redirect(w, r, "/", 302)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	if sess.Get("username") == nil {
		http.Redirect(w, r, "/login", 302)
	} else {
		user := sess.Get("username")
		countStr := fmt.Sprint(sess.Get("count"))
		count, _ := strconv.Atoi(countStr)
		sess.Set("count", count+1)
		fmt.Fprintf(w, "hi, %s! You have visited this page %d times.\n", user, count)
		fmt.Fprintf(w, "you are at host: %s\n", hostname)
	}
}

type myhandler struct{}

var loginPath = regexp.MustCompile("^/(login)")
var faviconPath = regexp.MustCompile("^/favicon\\.ico")

func (handler myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atLogin := loginPath.FindStringSubmatch(r.URL.Path)
	atFavicon := faviconPath.FindStringSubmatch(r.URL.Path)
	if atLogin != nil {
		login(w, r)
	} else if atFavicon != nil {
		fmt.Println("no favicon ")
	} else {
		index(w, r)
	}
}

func main() {
	//http.HandleFunc("/", index)
	//http.HandleFunc("/login", login)
	//endless.ListenAndServe(":8080", http.Handler)
	http.ListenAndServe(":8080", myhandler{})
}
