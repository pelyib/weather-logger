package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
	"go.etcd.io/bbolt"
)

const bucket string = "http" // Use the same bucket name as in the config [botond.pelyi]

type InMemmoryRepository struct {
	charts     map[string]*business.Chart
	originRepo business.ChartRepository
}

type DatabaseRepository struct {
	dbKey DatabaseKey
	db    bbolt.DB
	l     shared.Logger
}

type MigrationDatabaseRepository struct {
	r business.ChartRepository
	l shared.Logger
}

type DatabaseKey func(business.ChartSearchRequestI) []byte

func (repo InMemmoryRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	if _, ok := repo.charts[csr.GetYm()]; ok {
		return repo.charts[csr.GetYm()]
	}

	c := repo.originRepo.Load(csr)

	repo.charts[csr.GetYm()] = c

	return c
}

func (repo DatabaseRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	var c business.Chart = business.Chart{}
	err := repo.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		v := b.Get(repo.dbKey(csr))

		if v != nil {
			err := json.Unmarshal(v, &c)
			repo.l.Info("Chart found in database")
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
		repo.l.Info(fmt.Sprintf("Chart not found, creating empty for %s", csr.GetYm()))
		c = business.MakeEmptyChart(csr.GetYm())
	}

	return &c
}

func (r MigrationDatabaseRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	c := r.r.Load(csr)

	if c.IsNew && csr.HasLoc() {
		c = r.r.Load(csr.WithoutLoc())
	}

	return c
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
			b.Put(r.dbKey(business.ChartSearchRequest{Ym: c.Ym, Loc: c.Loc}), []byte(cjson))
			//b.Put([]byte(c.Ym), []byte(cjson))
		}

		return nil
	})

	if err != nil {
		r.l.Error(fmt.Sprintf("Saving failed, reason: %s", err.Error()))
	}
}

func (r MigrationDatabaseRepository) Save(c business.Chart) {
	r.r.Save(c)
}

func MakeChartRepository(db *bbolt.DB, l shared.Logger) business.ChartRepository {
	return InMemmoryRepository{
		charts: make(map[string]*business.Chart, 0),
		originRepo: DatabaseRepository{
			dbKey: func(csr business.ChartSearchRequestI) []byte {
				var key bytes.Buffer
				if csr.HasLoc() {
					key.WriteString(csr.GetLoc().Country().Alpha2Code())
					key.WriteString(csr.GetLoc().Name())
				}

				if csr.GetYm() != "" {
					key.WriteString(csr.GetYm())
				}

				return key.Bytes()
			},
			db: *db,
			l:  l,
		},
	}
}
