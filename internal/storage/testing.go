package storage

import (
	"fmt"
	"strings"
	"testing"
	"todolist/internal/config"
)

func TestStorage(t *testing.T) (*Storage, func(...string)) {
	t.Helper()

	cfg := config.MustLoad()
	s, err := New(cfg.Database.StoragePath)
	if err != nil {
		t.Fatal(err)
	}

	return s, func(tables ...string) {
		if len(tables) > 0 {
			if _, err := s.db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))); err != nil {
				t.Fatal(err)
			}
		}
		s.db.Close()
	}

}
