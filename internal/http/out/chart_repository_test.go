package out

import (
	"testing"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

// stubChartRepository is a test double for business.ChartRepository.
type stubChartRepository struct {
	charts map[string]*business.Chart
	saved  []business.Chart
}

func newStubRepo() *stubChartRepository {
	return &stubChartRepository{charts: make(map[string]*business.Chart)}
}

func (s *stubChartRepository) Load(csr business.ChartSearchRequestI) *business.Chart {
	key := csr.GetYm() + csr.GetLoc().Name
	if c, ok := s.charts[key]; ok {
		return c
	}
	empty := business.MakeEmptyChart(csr)
	return &empty
}

func (s *stubChartRepository) Save(c business.Chart) error {
	s.saved = append(s.saved, c)
	return nil
}

func TestInMemmoryRepository_LoadReturnsCachedChart(t *testing.T) {
	stub := newStubRepo()

	loc := shared.Location{Name: "Berlin"}
	loc.Country.Alpha2Code = "DE"

	key := func(csr business.ChartSearchRequestI) []byte {
		return []byte(csr.GetYm() + csr.GetLoc().Name)
	}

	repo := InMemmoryRepository{
		key:        key,
		charts:     make(map[string]*business.Chart),
		originRepo: stub,
	}

	csr := business.ChartSearchRequest{Ym: "2024-01", Loc: loc}

	// First load — goes to origin
	c1 := repo.Load(csr)
	if c1 == nil {
		t.Fatal("expected non-nil chart")
	}

	// Second load — should be served from cache (same pointer)
	c2 := repo.Load(csr)
	if c1 != c2 {
		t.Error("expected second load to return cached chart (same pointer)")
	}
}

func TestInMemmoryRepository_SaveUpdatesCache(t *testing.T) {
	stub := newStubRepo()

	key := func(csr business.ChartSearchRequestI) []byte {
		return []byte(csr.GetYm() + csr.GetLoc().Name)
	}

	repo := InMemmoryRepository{
		key:        key,
		charts:     make(map[string]*business.Chart),
		originRepo: stub,
	}

	loc := shared.Location{Name: "Berlin"}
	loc.Country.Alpha2Code = "DE"
	chart := business.MakeEmptyChart(business.ChartSearchRequest{Ym: "2024-01", Loc: loc})

	if err := repo.Save(chart); err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	if len(stub.saved) != 1 {
		t.Errorf("expected 1 save to origin repo, got %d", len(stub.saved))
	}
}
