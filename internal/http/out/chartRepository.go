package out

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
	"go.etcd.io/bbolt"
)

const bucket string = "charts.monthly"

type InMemmoryRepository struct {
	key        DatabaseKey
	charts     map[string]*business.Chart
	originRepo business.ChartRepository
}

type DatabaseRepository struct {
	dbKey DatabaseKey
	db    bbolt.DB
	l     shared.Logger
}

type DatabaseKey func(business.ChartSearchRequestI) []byte

func (repo InMemmoryRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	key := string(repo.key(csr))
	if _, ok := repo.charts[key]; ok {
		return repo.charts[key]
	}

	c := repo.originRepo.Load(csr)
	repo.charts[key] = c
	return c
}

func (repo DatabaseRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	var c business.Chart
	err := repo.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get(repo.dbKey(csr))
		if v == nil {
			return fmt.Errorf("chart not found")
		}
		repo.l.Info("Chart found in database")
		if err := json.Unmarshal(v, &c); err != nil {
			repo.l.Error(err.Error())
			return err
		}
		return nil
	})

	if err != nil {
		repo.l.Info(fmt.Sprintf("Chart not found, creating empty for %s", string(repo.dbKey(csr))))
		c = business.MakeEmptyChart(csr)
	}

	return &c
}

func (r InMemmoryRepository) Save(c business.Chart) error {
	r.charts[string(r.key(business.ChartSearchRequest{Ym: c.Ym, Loc: c.Loc}))] = &c
	return r.originRepo.Save(c)
}

func (r DatabaseRepository) Save(c business.Chart) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		cjson, err := json.Marshal(c)
		if err != nil {
			return err
		}
		r.l.Info("saving")
		return b.Put(r.dbKey(business.ChartSearchRequest{Ym: c.Ym, Loc: c.Loc}), cjson)
	})
}

func MakeChartRepository(db *bbolt.DB, l shared.Logger) business.ChartRepository {
	key := func(csr business.ChartSearchRequestI) []byte {
		var key bytes.Buffer
		if csr.HasLoc() {
			key.WriteString(csr.GetLoc().Country.Alpha2Code)
			key.WriteString(csr.GetLoc().Name)
		}
		if csr.GetYm() != "" {
			key.WriteString(csr.GetYm())
		}
		return key.Bytes()
	}

	return InMemmoryRepository{
		key:    key,
		charts: make(map[string]*business.Chart, 0),
		originRepo: DatabaseRepository{
			dbKey: key,
			db:    *db,
			l:     l,
		},
	}
}
