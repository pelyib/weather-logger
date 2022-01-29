package internal

import (
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

func MakeDb(cnf *LoggerCnf) bolt.DB {
  db, err := bolt.Open(fmt.Sprintf("%s%s", cnf.Database.Folder, "0.2.0.db"), 0600, nil)
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
