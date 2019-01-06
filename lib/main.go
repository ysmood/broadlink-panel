package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/tidwall/gjson"
	g "github.com/ysmood/gokit"
	"golang.org/x/oauth2"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type server struct {
	dev   *device
	oauth *oauth2.Config
	store *sessions.CookieStore
}

var (
	host     = kingpin.Flag("host", "").Short('h').Default("http://localhost:7382").String()
	port     = kingpin.Flag("port", "").Short('p').Default("7382").String()
	accounts = kingpin.Flag("accounts", "email address list").Short('a').Strings()
	clientID = kingpin.Flag("client-id", "google app client id").Short('c').String()
	secret   = kingpin.Flag("secret", "").Short('s').String()
)

func main() {
	kingpin.Parse()

	dev, err := newDevice()
	if err != nil {
		g.Err(err)
	}

	s := &server{
		dev: dev,
		oauth: &oauth2.Config{
			ClientID:     *clientID,
			ClientSecret: *secret,
			Scopes:       []string{"https://www.googleapis.com/auth/gmail.metadata"},
			RedirectURL:  *host + "/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
		store: sessions.NewCookieStore([]byte(*secret)),
	}

	http.HandleFunc("/", s.handler)

	g.Log("listen on:", *port)
	g.E(http.ListenAndServe(":"+*port, nil))
}

func (s *server) handler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			g.Err(r)

			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprint(r)))
		}
	}()

	args := strings.Split(req.URL.Path[1:], "/")

	if args[0] != "login" && args[0] != "cb" {
		if !s.isLogin(req) {
			w.Header().Set("Location", "/login?backto="+url.QueryEscape(req.URL.Path))
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}

	switch args[0] {
	case "login":
		s.login(w, req)
		return

	case "cb":
		s.loginCallback(w, req)
		return

	case "learn":
		g.E(s.dev.learn(args[1]))
		s.home(w)

	case "group":
		g.E(s.dev.addGroup(args[1], args[2:]))
		s.home(w)

	case "send":
		g.E(s.dev.send(args[1]))
		s.home(w)

	case "list":
		l, err := s.dev.list()
		g.E(err)

		data, err := json.Marshal(l)
		g.E(err)

		w.Write(data)
		return

	case "rename":
		g.E(s.dev.rename(args[1], args[2]))
		s.home(w)

	case "delete":
		g.E(s.dev.delete(args[1]))
		s.home(w)

	default:
		http.FileServer(http.Dir("web")).ServeHTTP(w, req)
		return
	}

	w.Write([]byte("OK"))
}

func (s *server) login(w http.ResponseWriter, req *http.Request) {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := s.oauth.AuthCodeURL(req.URL.Query().Get("backto"), oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *server) loginCallback(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	code := req.URL.Query().Get("code")

	tok, err := s.oauth.Exchange(ctx, code)
	g.E(err)

	client := s.oauth.Client(ctx, tok)
	res, err := client.Get("https://www.googleapis.com/gmail/v1/users/me/profile")
	g.E(err)

	body, err := ioutil.ReadAll(res.Body)
	g.E(err)

	email := gjson.Get(string(body), "emailAddress").String()

	if !has(*accounts, email) {
		panic("you are not in the admin account list")
	}

	session, err := s.store.Get(req, "token")
	g.E(err)
	session.Options.MaxAge = int((365 * 24 * time.Hour).Seconds())
	session.Values["login"] = true
	session.Save(req, w)

	w.Header().Set("Location", req.URL.Query().Get("state"))
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *server) home(w http.ResponseWriter) {
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *server) isLogin(r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token != "" {
		g.Log(token)
		r.Header.Set("Cookie", token)
	}

	session, err := s.store.Get(r, "token")
	g.E(err)

	_, login := session.Values["login"]

	return login
}

func has(list []string, str string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}
