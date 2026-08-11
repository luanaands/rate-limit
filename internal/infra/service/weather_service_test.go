package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWeather(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "fakekey" || r.URL.Query().Get("q") != "São Paulo" {
			t.Fatalf("unexpected query params: key=%s, q=%s", r.URL.Query().Get("key"), r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
  "location": {
    "name": "Sao Paulo",
    "region": "Sao Paulo",
    "country": "Brazil",
    "lat": -23.5333,
    "lon": -46.6167,
    "tz_id": "America/Sao_Paulo",
    "localtime_epoch": 1776481265,
    "localtime": "2026-04-18 00:01"
  },
  "current": {
    "last_updated_epoch": 1776481200,
    "last_updated": "2026-04-18 00:00",
    "temp_c": 21.3,
    "temp_f": 70.3,
    "is_day": 0,
    "condition": {
      "text": "Partly Cloudy",
      "icon": "//cdn.weatherapi.com/weather/64x64/night/116.png",
      "code": 1003
    },
    "wind_mph": 4.3,
    "wind_kph": 6.8,
    "wind_degree": 338,
    "wind_dir": "NNW",
    "pressure_mb": 1012,
    "pressure_in": 29.88,
    "precip_mm": 0,
    "precip_in": 0,
    "humidity": 78,
    "cloud": 0,
    "feelslike_c": 21.3,
    "feelslike_f": 70.3,
    "windchill_c": 20.5,
    "windchill_f": 68.9,
    "heatindex_c": 20.5,
    "heatindex_f": 68.9,
    "dewpoint_c": 13,
    "dewpoint_f": 55.5,
    "vis_km": 10,
    "vis_miles": 6,
    "uv": 0,
    "gust_mph": 8.3,
    "gust_kph": 13.3,
    "short_rad": 0,
    "diff_rad": 0,
    "dni": 0,
    "gti": 0
  }
}`)
	}))
	defer ts.Close()

	s := &WeatherService{client: ts.Client()}
	resp, err := s.GetWeather("São Paulo", "fakekey", ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, resp.TempC, 21.3)
	assert.Equal(t, resp.TempF, 70.3)
}
