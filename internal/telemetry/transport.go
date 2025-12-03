package telemetry

import (
	"log"
	"net/http"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/config"
)

type MetricTransport struct {
	Base    http.RoundTripper    // The original transporter (usually http.DefaultTransport)
	DB      *config.DBConnection // tsnet/postgres connection wrapper
	Service string               // "steam", "spotify", "github"
}

func (t *MetricTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	db := t.DB.GetDB() // we either have it or we dont

	if db == nil {
		resp, err := t.Base.RoundTrip(req)
		log.Printf("[METRICS] metrics are either disabled or still connecting, however this request to %v took %v", t.Service, time.Since(start))

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
		_, dbErr := db.Exec(query, req.URL.String(), req.Method, statusCode, duration.Milliseconds(), t.Service)
		log.Printf("[METRICS] Trying to write to db :: %v @ %vms", t.Service, duration.Milliseconds())

		if dbErr != nil {

			// don't crash
			log.Printf("[METRICS] DB Log Error: %v\n", dbErr)
		}
	}()

	return resp, err
}
