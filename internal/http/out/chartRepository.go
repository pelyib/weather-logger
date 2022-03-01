package out

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/pelyib/weather-logger/internal/http/business"
	"go.etcd.io/bbolt"
)

const bucket string = "http"

type InMemmoryRepository struct {
  charts map[string]*business.Chart
  originRepo business.ChartRepository
}

type DatabaseRepository struct {
  db bbolt.DB
}

func (repo InMemmoryRepository) Load(searchRequest business.ChartSearchRequest) *business.Chart {
  if _, ok := repo.charts[searchRequest.Ym]; ok {
    return repo.charts[searchRequest.Ym]
  }

  c := repo.originRepo.Load(searchRequest)

  repo.charts[searchRequest.Ym] = c

  return c
}

func (repo DatabaseRepository) Load(searchRequest business.ChartSearchRequest) *business.Chart {
  var c business.Chart = business.Chart{}
  err := repo.db.View(func(tx *bbolt.Tx) error {
    b := tx.Bucket([]byte(bucket))

    v := b.Get([]byte(searchRequest.Ym))

    if v != nil {
      err := json.Unmarshal(v, &c)
      log.Println("ChartRepository: chart found in DB")
      //log.Println(string(v[:]))
      if err != nil {
        log.Println(err)
        return err
      } else {
        return nil
      }
    }

    return errors.New("Chart not foud")
  })

  if err != nil {
    log.Println(fmt.Sprintf("ChartRepository : Chart not found, creating empty for %s", searchRequest.Ym))
    c = business.MakeEmptyChart(searchRequest.Ym)
  }

  return &c
}

func (r InMemmoryRepository) Save(c business.Chart) {
  r.charts[c.Ym] = &c

  r.originRepo.Save(c)
}

func (r DatabaseRepository) Save(c business.Chart) {
  err := r.db.Update(func(tx *bbolt.Tx) error {
    b := tx.Bucket([]byte(bucket))

    if cjson, err := json.Marshal(c); err != nil {
      return err
    } else {
      log.Println("ChartRepository: saving")
      log.Println(string(cjson))
      b.Put([]byte(c.Ym), []byte(cjson))
    }

    return nil
  })

  if err != nil {
    log.Println("chartRepository: saving failed")
    log.Println(err)
  }
}

func MakeChartRepository(db *bbolt.DB) business.ChartRepository {
  return InMemmoryRepository{
    charts: make(map[string]*business.Chart, 0),
    originRepo: DatabaseRepository{
      db: *db,
    },
  }
}
