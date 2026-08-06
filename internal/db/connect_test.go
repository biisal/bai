package db

import (
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := Connect("file::memory:?_fk=1")
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() failed: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() failed: %v", err)
	}
}
