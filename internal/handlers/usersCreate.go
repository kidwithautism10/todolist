package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleUsersCreate() http.HandlerFunc {
	const op = "hanlers/usersCreate"
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			fmt.Errorf("%s : %w , status: %s", op, err, http.StatusBadRequest)
			return
		}
	}
}
