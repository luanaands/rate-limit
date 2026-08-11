package service

import (
	"github.com/luanaands/rate-limit/internal/dto"
)

type CepInterface interface {
	GetViaCep(cep string, url string) (*dto.CepResponse, error)
}

type WeatherInterface interface {
	GetWeather(city string, apiKey string, baseURL string) (*dto.WeatherResponse, error)
}
