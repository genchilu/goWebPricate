package main

import (
	"fmt"
	_ "github.com/genchilu/goWebPricate/memory"
	"github.com/genchilu/goWebPricate/session"
	"html/template"
	"net/http"
	"strconv"
)

var globalSessions *session.Manager

// Then, initialize the session manager
func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	//go globalSessions.GC()
	fmt.Println("finish init main")
}

func login(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	fmt.Println(sess)
	r.ParseForm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, sess.Get("username"))
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
		fmt.Fprintf(w, "hi, %s! You have visited this page %d times.", user, count)
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.ListenAndServe(":8080", nil)
}
