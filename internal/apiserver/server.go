package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"net/http"
	"todolist/internal/storage"
)

type server struct {
	router       *chi.Mux
	store        storage.Store
	sessionStore sessions.Store
}

const SessionName = "session0"

var (
	errIncorrectUsernameOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated            = errors.New("not authenticated")
)

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
