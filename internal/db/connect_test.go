package db

import (
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := Connect("file::memory:?_fk=1")
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() failed: %v", err)
	}
}
