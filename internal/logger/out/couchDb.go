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

type client struct {
	config shared.CouchDb
	now    func() time.Time
}

type apiRawResponses struct {
	CalledAt string      `json:"calledAt"`
	SourceId string      `json:"sourceId"`
	Raw      interface{} `json:"raw"`
}

type apiRawResponsesDbSchema struct {
	Obj apiRawResponses `json:"obj"`
}

func (c client) saveRawApiRes(sourceId string, rawApiRes []byte) error {
	var result interface{}
	err := json.Unmarshal(rawApiRes, &result)
	if err != nil {
		return err
	}

	couchdbReqBody := apiRawResponsesDbSchema{
		Obj: apiRawResponses{
			CalledAt: c.now().Format(time.RFC3339),
			SourceId: sourceId,
			Raw:      result,
		},
	}

	couchdbSerializedRecord, _ := json.Marshal(couchdbReqBody)

	return c.put("api_raw_responses", couchdbSerializedRecord)
}

func (c client) saveMeasurements(measurements []shared.MeasurementResult) error {
	for _, measurement := range measurements {
		data := struct {
			Obj shared.MeasurementResult `json:"obj"`
		}{Obj: measurement}

		serializedMeasurement, _ := json.Marshal(data)

		err := c.put("measurements", serializedMeasurement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c client) put(dbId string, data []byte) error {
	var selectedDb shared.Db
	for id, db := range c.config.Dbs {
		if id == dbId {
			selectedDb = db
		}
	}
	if (selectedDb == shared.Db{}) {
		return errors.New(fmt.Sprintf("No DB config specified for %s", dbId))
	}

	rand.Seed(c.now().UnixNano())
	putReq, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/%s/%d", c.config.Host, selectedDb.Name, rand.Intn(10000)),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	putReq.Header.Add("Accept", "application/json")
	putReq.SetBasicAuth(selectedDb.User, selectedDb.Password)

	client := http.Client{}
	res, err := client.Do(putReq)

	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Could not save, status code: %d", res.StatusCode))
	}

	return nil
}

func MakeNewClient(conf shared.CouchDb) client {
	return client{
		now:    now,
		config: conf,
	}
}
