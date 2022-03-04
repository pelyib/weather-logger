package internal

import (
	"fmt"

	"github.com/pelyib/weather-logger/internal/shared"
	bolt "go.etcd.io/bbolt"
)

func MakeDb(cnf *shared.Database, l shared.Logger) bolt.DB {
	db, err := bolt.Open(fmt.Sprintf("%s%s", cnf.Folder, cnf.FileName), 0600, nil)
	if err != nil {
		l.Error(err.Error())
	}

	db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range cnf.Buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))

			if err != nil {
				l.Error(err.Error())
			}
		}

		return nil
	})

	return *db
}
