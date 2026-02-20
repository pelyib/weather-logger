package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pelyib/weather-logger/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS measurements (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	source           TEXT    NOT NULL,
	type             TEXT    NOT NULL CHECK(type IN ('forecast', 'historical')),
	min              REAL    NOT NULL,
	max              REAL    NOT NULL,
	measured_at      TEXT    NOT NULL,
	recorded_at      TEXT    NOT NULL,
	location_name    TEXT    NOT NULL,
	location_country TEXT    NOT NULL,
	location_lat     REAL    NOT NULL,
	location_lon     REAL    NOT NULL,
	location_aw_key  TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_measurements_month_loc
    ON measurements (strftime('%Y-%m', measured_at), location_name, location_country);
`

// Repository is a SQLite-backed MeasurementRepository.
type Repository struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite database at path and applies the schema.
func Open(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %q: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("sqlite: apply schema: %w", err)
	}
	return &Repository{db: db}, nil
}

// Close closes the underlying database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// Save persists a single Measurement.
func (r *Repository) Save(m domain.Measurement) error {
	_, err := r.db.Exec(`
		INSERT INTO measurements
			(source, type, min, max, measured_at, recorded_at,
			 location_name, location_country, location_lat, location_lon, location_aw_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Source,
		string(m.Type),
		m.Min,
		m.Max,
		m.At.UTC().Format(time.RFC3339),
		m.RecordedAt.UTC().Format(time.RFC3339),
		m.Location.Name,
		m.Location.Country.Alpha2Code,
		m.Location.Lat,
		m.Location.Lon,
		m.Location.AccuWeatherKey,
	)
	if err != nil {
		return fmt.Errorf("sqlite: save measurement: %w", err)
	}
	return nil
}

// FindByMonthAndLocation returns all measurements for a given year/month and
// location, ordered by measured_at ascending.
func (r *Repository) FindByMonthAndLocation(year int, month time.Month, loc domain.Location) ([]domain.Measurement, error) {
	ym := fmt.Sprintf("%04d-%02d", year, int(month))

	rows, err := r.db.Query(`
		SELECT source, type, min, max, measured_at, recorded_at
		FROM measurements
		WHERE strftime('%Y-%m', measured_at) = ?
		  AND location_name    = ?
		  AND location_country = ?
		ORDER BY measured_at ASC`,
		ym, loc.Name, loc.Country.Alpha2Code,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: query measurements: %w", err)
	}
	defer rows.Close()

	var results []domain.Measurement
	for rows.Next() {
		var (
			source     string
			mType      string
			min, max   float64
			measuredAt string
			recordedAt string
		)
		if err := rows.Scan(&source, &mType, &min, &max, &measuredAt, &recordedAt); err != nil {
			return nil, fmt.Errorf("sqlite: scan row: %w", err)
		}
		at, err := time.Parse(time.RFC3339, measuredAt)
		if err != nil {
			return nil, fmt.Errorf("sqlite: parse measured_at %q: %w", measuredAt, err)
		}
		rec, err := time.Parse(time.RFC3339, recordedAt)
		if err != nil {
			return nil, fmt.Errorf("sqlite: parse recorded_at %q: %w", recordedAt, err)
		}
		results = append(results, domain.Measurement{
			Source:     source,
			Type:       domain.MeasurementType(mType),
			Min:        min,
			Max:        max,
			At:         at,
			RecordedAt: rec,
			Location:   loc,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate rows: %w", err)
	}
	return results, nil
}
