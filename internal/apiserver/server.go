package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"html/template"
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

	s.router.Mount("/todo", s.loggedRouter())
}

func (s *server) loggedRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(s.authenticateUser)
	r.MethodFunc("POST", "/create", s.handleTodoCreate())
	r.MethodFunc("POST", "/update", s.handleCompleteTask())
	r.MethodFunc("POST", "/delete", s.handleDeleteTask())
	r.MethodFunc("GET", "/", s.handleRenderTask())
	return r
}

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateParser, err := template.ParseFiles("views/index.html")
		if err != nil {
			s.error(w, r, http.StatusNotFound, err)
		}

		templateParser.ExecuteTemplate(w, "index", nil)
	}
}

func (s *server) handleRenderTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := s.store.User().Find(r.Context().Value(ctxKeyUser).(*storage.User).ID)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
		}

		t, err := s.store.Task().RenderTask(u.Username)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
		fmt.Println(t)
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
		fmt.Println(req.Text)
		date := strconv.Itoa(time.Now().Day()) + "." + strconv.Itoa(int(time.Now().Month())) + "." + strconv.Itoa(time.Now().Year())

		u, err := s.store.User().Find(r.Context().Value(ctxKeyUser).(*storage.User).ID)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
		}

		if err := s.store.Task().CreateTask(req.Text, date, u.Username); err != nil {
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
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
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
