package telemetry

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

type MetricTransport struct {
	Base    http.RoundTripper // The original transporter (usually http.DefaultTransport)
	DB      *sql.DB           // tsnet/postgres connection
	Service string            // "steam", "spotify", "github"
}

func (t *MetricTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	if t.DB == nil {
		resp, err := t.Base.RoundTrip(req)
		log.Printf("Logging metrics are disabled, however this request to %v took %v", t.Service, time.Since(start))

		return resp, err
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	resp, err := base.RoundTrip(req)

	duration := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	go func() {
		query := `
			INSERT INTO api_metrics (target_url, method, status_code, duration_ms, service_name) 
			VALUES ($1, $2, $3, $4, $5)
		`
		_, dbErr := t.DB.Exec(query, req.URL.String(), req.Method, statusCode, duration.Milliseconds(), t.Service)
		if dbErr != nil {

			// don't crash
			log.Printf("DB Log Error: %v\n", dbErr)
		}
	}()

	return resp, err
}
