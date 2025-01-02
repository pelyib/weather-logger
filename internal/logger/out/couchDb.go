package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type now func() time.Time

type client struct {
	config shared.CouchDb
	now    now
}

type record struct {
	CalledAt string                 `json:"calledAt"`
	SourceId string                 `json:"sourceId"`
	Raw      map[string]interface{} `json:"raw"`
}

type dbSchema struct {
	Obj record `json:"obj"`
}

func (c client) saveRawApiRes(sourceId string, rawApiRes []byte) error {
	var selectedDb shared.Db
	for id, db := range c.config.Dbs {
		if id == "api_raw_responses" {
			selectedDb = db
		}
	}
	if (selectedDb == shared.Db{}) {
		return errors.New("No DB config specified for api_raw_responses")
	}

	var result map[string]interface{}
	err := json.Unmarshal(rawApiRes, &result)
	if err != nil {
		return err
	}

	couchdbReqBody := dbSchema{
		Obj: record{
			CalledAt: c.now().Format(time.RFC3339),
			SourceId: sourceId,
			Raw:      result,
		},
	}

	couchdbSerializedRecord, _ := json.Marshal(couchdbReqBody)

	rand.Seed(c.now().UnixNano())
	addToCouchDBReq, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/%s/%d", c.config.Host, selectedDb.Name, rand.Intn(10000)),
		bytes.NewBuffer(couchdbSerializedRecord),
	)
	if err != nil {
		return err
	}
	addToCouchDBReq.Header.Add("Accept", "application/json")
	addToCouchDBReq.SetBasicAuth("logger", "logger")

	client := http.Client{}
	couchDbRes, err := client.Do(addToCouchDBReq)

	if err != nil {
		return err
	}

	if couchDbRes.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Could not save, status code: %d", couchDbRes.StatusCode))
	}

	return nil
}

func MakeNewClient(conf shared.CouchDb) client {
	return client{
		now: func() time.Time {
			return time.Now()
		},
		config: conf,
	}
}
