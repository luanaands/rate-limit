package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luanaands/rate-limit/internal/dto"
	"github.com/stretchr/testify/assert"
)

// Mock implementations
type mockCepService struct {
	response *dto.CepResponse
	err      error
}

func (m *mockCepService) GetViaCep(cep, url string) (*dto.CepResponse, error) {
	return m.response, m.err
}

type mockWeatherService struct {
	response *dto.WeatherResponse
	err      error
}

func (m *mockWeatherService) GetWeather(city, key, host string) (*dto.WeatherResponse, error) {
	return m.response, m.err
}

func TestGetCep_MissingCep(t *testing.T) {
	handler := &CepHandler{
		Service:        &mockCepService{},
		WeatherService: &mockWeatherService{},
	}
	req := httptest.NewRequest("GET", "/weather", nil)
	ctx := context.WithValue(context.WithValue(context.WithValue(req.Context(), "ViaCepHost", "http://mock"), "ApiWeatherHost", "http://mock"), "ApiWeatherKey", "key")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCep(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "CEP é obrigatório", resp["error"])
}

func TestGetCep_InvalidCepLength(t *testing.T) {
	handler := &CepHandler{
		Service:        &mockCepService{},
		WeatherService: &mockWeatherService{},
	}
	req := httptest.NewRequest("GET", "/weather?cep=123", nil)
	ctx := context.WithValue(context.WithValue(context.WithValue(req.Context(), "ViaCepHost", "http://mock"), "ApiWeatherHost", "http://mock"), "ApiWeatherKey", "key")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCep(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "invalid zipcode", resp["error"])
}

func TestGetCep_GetViaCepError(t *testing.T) {
	handler := &CepHandler{
		Service:        &mockCepService{err: assert.AnError},
		WeatherService: &mockWeatherService{},
	}
	req := httptest.NewRequest("GET", "/weather?cep=01001000", nil)
	ctx := context.WithValue(context.WithValue(context.WithValue(req.Context(), "ViaCepHost", "http://mock"), "ApiWeatherHost", "http://mock"), "ApiWeatherKey", "key")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCep(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "can not find zipcode", resp["error"])
}

func TestGetCep_GetWeatherError(t *testing.T) {
	viaCepResp := &dto.CepResponse{Localidade: "São Paulo"}
	handler := &CepHandler{
		Service:        &mockCepService{response: viaCepResp},
		WeatherService: &mockWeatherService{err: assert.AnError},
	}
	req := httptest.NewRequest("GET", "/weather?cep=01001000", nil)
	ctx := context.WithValue(context.WithValue(context.WithValue(req.Context(), "ViaCepHost", "http://mock"), "ApiWeatherHost", "http://mock"), "ApiWeatherKey", "key")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCep(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "can not find weather", resp["error"])
}

func TestGetCep_Success(t *testing.T) {
	viaCepResp := &dto.CepResponse{Localidade: "São Paulo"}
	weatherResp := &dto.WeatherResponse{
		TempC: 25.0,
		TempF: 77.0,
	}
	handler := &CepHandler{
		Service:        &mockCepService{response: viaCepResp},
		WeatherService: &mockWeatherService{response: weatherResp},
	}
	req := httptest.NewRequest("GET", "/weather?cep=01001000", nil)
	ctx := context.WithValue(context.WithValue(context.WithValue(req.Context(), "ViaCepHost", "http://mock"), "ApiWeatherHost", "http://mock"), "ApiWeatherKey", "key")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCep(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 25.0, resp.TempC)
	assert.Equal(t, 77.0, resp.TempF)
	assert.Equal(t, 298.0, resp.TempK)
}
