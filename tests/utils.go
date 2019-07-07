package tests

import (
	"github.com/cybersamx/goa"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"testing"
)

func setupDB(t *testing.T) (*gorm.DB, func()) {
	dsn := ":memory:"
	db, err := gorm.Open("sqlite3", dsn)
	closeFunc := func() {
		if db != nil {
			log.Printf("closing database.\n")
			csModel := goa.ClientStoreItem{}
			db.DropTableIfExists(&csModel)
			db.Close()
		}
	}

	if err != nil {
		t.Fatalf("problem opening %s, details: %v", dsn, err)
		return nil, closeFunc
	}

	// Configure the pool
	db.DB().SetMaxOpenConns(1)
	db.DB().SetMaxIdleConns(1)

	return db, closeFunc
}