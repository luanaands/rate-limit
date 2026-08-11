package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/luanaands/rate-limit/internal/dto"
	"github.com/luanaands/rate-limit/internal/entity"
)

type WeatherService struct {
	client *http.Client
}

func NewWeatherService() *WeatherService {
	return &WeatherService{
		client: &http.Client{},
	}
}

func (s *WeatherService) GetWeather(city string, apiKey string, baseURL string) (*dto.WeatherResponse, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", apiKey)
	q.Set("q", city)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response entity.WeatherResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return dto.FromWeather(&response), nil
}
