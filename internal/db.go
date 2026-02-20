package internal

import (
	"fmt"

	"github.com/pelyib/weather-logger/internal/shared"
	bolt "go.etcd.io/bbolt"
)

func MakeDb(cnf *shared.Database, l shared.Logger) (*bolt.DB, error) {
	db, err := bolt.Open(fmt.Sprintf("%s%s", cnf.Folder, cnf.FileName), 0600, nil)
	if err != nil {
		l.Error(err.Error())
		return nil, fmt.Errorf("open database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range cnf.Buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("create bucket %q: %w", bucket, err)
			}
		}
		return nil
	})

	if err != nil {
		l.Error(err.Error())
		return nil, fmt.Errorf("initialize buckets: %w", err)
	}

	return db, nil
}
