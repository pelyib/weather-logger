package internal

import (
	"fmt"
	"log"

	"github.com/pelyib/weather-logger/internal/shared"
	bolt "go.etcd.io/bbolt"
)

func MakeDb(cnf *shared.Database) bolt.DB {
  db, err := bolt.Open(fmt.Sprintf("%s%s", cnf.Folder, cnf.FileName), 0600, nil)
  if err != nil {
    log.Fatal(err)
  }

  db.Update(func(tx *bolt.Tx) error {
    tx.CreateBucketIfNotExists([]byte("http")) // TODO: fetch it from Config

    // TODO: create another buckets
    // TODO: check errors

    return nil
  })

  return *db
}
