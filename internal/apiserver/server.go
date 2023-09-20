package apiserver

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"todolist/internal/storage"
	"todolist/internal/storage/sqlstore"
)

type server struct {
	router *chi.Mux
	store  *sqlstore.Storage
}

func newServer(store *sqlstore.Storage) *server {
	s := &server{
		router: chi.NewRouter(),
		store:  store,
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

	s.router.Handle("/user/create", s.HandleUsersCreate())
}

func (s *server) HandleUsersCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	const op = "handlers/usersCreate"
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			fmt.Printf("%s : %w , status: %s", op, err, http.StatusBadRequest)
			return
		}

		u := &storage.User{
			Username: req.Username,
			Password: req.Password,
		}
		fmt.Println(u.Username, u.Password)
		if err := s.store.User().Create(u); err != nil {
			fmt.Errorf("%w, status code: %d", err, http.StatusUnprocessableEntity)
			return
		}

		fmt.Println(u.ID)

		fmt.Printf("%d", http.StatusCreated)

	}
}
