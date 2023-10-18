package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"todolist/internal/storage"
)

type server struct {
	router       *chi.Mux
	store        storage.Store
	sessionStore sessions.Store
}

const (
	SessionName        = "session0"
	ctxKeyUser  ctxKey = iota
)

var (
	errIncorrectUsernameOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated            = errors.New("not authenticated")
)

type ctxKey int8

func newServer(store storage.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       chi.NewRouter(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	fs := http.FileServer(http.Dir("views/static/"))
	s.router.Handle("/static/*", http.StripPrefix("/static", fs))

	s.router.MethodFunc("POST", "/user/create", s.handleUsersCreate())
	s.router.MethodFunc("POST", "/sessions", s.handleSessionsCreate())
	s.router.Handle("/", s.handleIndex())
	s.router.Handle("/auth", s.handleAuth())

	s.router.Mount("/todo", s.loggedRouter())
}

func (s *server) loggedRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(s.authenticateUser)
	r.MethodFunc("POST", "/create", s.handleTodoCreate())
	r.MethodFunc("POST", "/update", s.handleCompleteTask())
	r.MethodFunc("DELETE", "/delete", s.handleDeleteTask())
	r.MethodFunc("GET", "/", s.handleRenderTask())
	r.MethodFunc("GET", "/json", s.handleJSON())
	r.Handle("/weather", s.handleWeather())
	r.MethodFunc("POST", "/getw", s.handleGetWeather())
	return r
}

func (s *server) handleJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dat, _ := ioutil.ReadFile("./views/static/countries-110m.json")
		w.Write(dat)
	}
}

func (s *server) handleWeather() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateParser, err := template.ParseFiles("views/weather.html")
		if err != nil {
			s.error(w, r, http.StatusNotFound, err)
		}

		templateParser.ExecuteTemplate(w, "weather", nil)
	}
}

func (s *server) handleGetWeather() http.HandlerFunc {
	type request struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	type response struct {
		Temp      string `json:"temp"`
		Condition string `json:"condition"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		url := "https://ru.meteotrend.com/forecast/" + req.Country + "/" + req.City + "/"
		fmt.Println(url)
		res, err := http.Get(url)
		if err != nil {
			return
		}
		defer res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(res.Body)
		condititon, _ := doc.Find("div.box:nth-child(2) > a:nth-child(1) > div:nth-child(2) > div:nth-child(1) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(2) > img:nth-child(1)").Attr("alt")
		respon := &response{
			Temp:      doc.Find("div.box:nth-child(2) > a:nth-child(1) > div:nth-child(2) > div:nth-child(1) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(2) > b:nth-child(2)").Text(),
			Condition: condititon,
		}
		s.respond(w, r, http.StatusOK, respon)
	}
}

func (s *server) handleAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session0")
		if err != nil {
			templateParser, err := template.ParseFiles("views/auth.html")
			if err != nil {
				s.error(w, r, http.StatusNotFound, err)
			}

			templateParser.ExecuteTemplate(w, "auth", nil)
		} else {
			http.Redirect(w, r, "/todo", http.StatusSeeOther)
		}

	}
}

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session0")
		if err != nil {
			templateParser, err := template.ParseFiles("views/index.html")
			if err != nil {
				s.error(w, r, http.StatusNotFound, err)
			}

			templateParser.ExecuteTemplate(w, "index", nil)
		} else {
			http.Redirect(w, r, "/todo", http.StatusSeeOther)
		}

	}
}

func (s *server) handleRenderTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateParser, err := template.ParseFiles("views/main.html")
		if err != nil {
			s.error(w, r, http.StatusNotFound, err)
		}
		u, err := s.store.User().Find(r.Context().Value(ctxKeyUser).(*storage.User).ID)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
		}

		t, err := s.store.Task().RenderTask(u.Username)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
		templateParser.ExecuteTemplate(w, "main", t)
	}
}

func (s *server) handleDeleteTask() http.HandlerFunc {
	type request struct {
		Id string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		id, _ := strconv.Atoi(req.Id)

		if err := s.store.Task().DeleteTask(id); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (s *server) handleCompleteTask() http.HandlerFunc {
	type request struct {
		Id string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		id, _ := strconv.Atoi(req.Id)

		if err := s.store.Task().CompleteTask(id); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (s *server) handleTodoCreate() http.HandlerFunc {
	type request struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		date := strconv.Itoa(time.Now().Day()) + "." + strconv.Itoa(int(time.Now().Month())) + "." + strconv.Itoa(time.Now().Year())

		u, err := s.store.User().Find(r.Context().Value(ctxKeyUser).(*storage.User).ID)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
		}
		t := &storage.Task{
			Text:     req.Text,
			Date:     date,
			Username: u.Username,
		}
		if err := s.store.Task().CreateTask(t); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
	}
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, SessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &storage.User{
			Username: req.Username,
			Password: req.Password,
		}
		fmt.Println(u.Username, u.Password)

		userCreated, err := s.store.User().FindByUsername(u.Username)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
		if userCreated == nil {
			if err := s.store.User().Create(u); err != nil {
				s.error(w, r, http.StatusUnprocessableEntity, err)
				return
			}
		} else {
			s.error(w, r, http.StatusBadRequest, errors.New("Username already taken!"))
		}

		fmt.Println(u.ID)

		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByUsername(req.Username)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectUsernameOrPassword)
			return
		}

		session, err := s.sessionStore.Get(r, SessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = u.ID

		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
