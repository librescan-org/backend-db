package database

import (
	"log"
	"time"

	storage "github.com/librescan-org/backend-db"
	"github.com/librescan-org/backend-db/database/postgres"
)

func LoadRepository() storage.Storage {
	repo := &postgres.PostgresRepository{}
	for {
		err := repo.Load()
		if err == nil {
			break
		} else {
			log.Printf("failed to initialize database: %v", err)
			time.Sleep(time.Second * 1)
		}
	}
	return repo
}
