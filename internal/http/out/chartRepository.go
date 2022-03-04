package out

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
	"go.etcd.io/bbolt"
)

const bucket string = "http"

type InMemmoryRepository struct {
	charts     map[string]*business.Chart
	originRepo business.ChartRepository
}

type DatabaseRepository struct {
	db bbolt.DB
	l  shared.Logger
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
			repo.l.Info("Chart found in datebase")
			if err != nil {
				repo.l.Error(err.Error())
				return err
			} else {
				return nil
			}
		}

		return errors.New("Chart not foud")
	})

	if err != nil {
		repo.l.Info(fmt.Sprintf("Chart not found, creating empty for %s", searchRequest.Ym))
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
			r.l.Info("saving")
			b.Put([]byte(c.Ym), []byte(cjson))
		}

		return nil
	})

	if err != nil {
		r.l.Error(fmt.Sprintf("Saving failed, reason: %s", err.Error()))
	}
}

func MakeChartRepository(db *bbolt.DB, l shared.Logger) business.ChartRepository {
	return InMemmoryRepository{
		charts: make(map[string]*business.Chart, 0),
		originRepo: DatabaseRepository{
			db: *db,
			l:  l,
		},
	}
}
