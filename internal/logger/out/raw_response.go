package out

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
	bolt "go.etcd.io/bbolt"
)

const (
	bucketAccuWeather = "accuweather.raw_response"
	bucketOpenWeather = "openweather.raw_response"
	bucketOpenMeteo   = "openmeteo.raw_response"
)

// saveRawResponse writes body into bucket under a time-prefixed key.
// It is a no-op when db is nil, so providers work safely in tests without a DB.
func saveRawResponse(db *bolt.DB, bucket, locName string, body []byte, l shared.Logger) {
	if db == nil {
		return
	}
	key := fmt.Sprintf("%s:%s", time.Now().UTC().Format(time.RFC3339), locName)
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		return b.Put([]byte(key), body)
	}); err != nil {
		l.Error(fmt.Sprintf("Failed to save raw response to %q: %s", bucket, err.Error()))
	}
}
