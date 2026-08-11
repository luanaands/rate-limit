package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/luanaands/rate-limit/internal/dto"
	"github.com/luanaands/rate-limit/internal/infra/service"
)

type CepHandler struct {
	Service        service.CepInterface
	WeatherService service.WeatherInterface
}

func NewCepHandler(service service.CepInterface, weatherService service.WeatherInterface) *CepHandler {
	return &CepHandler{
		Service:        service,
		WeatherService: weatherService,
	}
}

// @Summary Buscar clima atual
// @Description Retorna dados do tempo consultando ViaCEP e WeatherAPI
// @Tags CEP
// @Accept json
// @Produce json
// @Param cep query string true "CEP sem formatação (ex: 01001000)"
// @Router /weather [get]
func (h *CepHandler) GetCep(w http.ResponseWriter, r *http.Request) {
	viaCepUrl := r.Context().Value("ViaCepHost").(string)
	apiWeatherHost := r.Context().Value("ApiWeatherHost").(string)
	apiWeatherKey := r.Context().Value("ApiWeatherKey").(string)
	cep := r.URL.Query().Get("cep")

	if cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "CEP é obrigatório"})
		return
	}

	if len(cep) != 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	viaCepResponse, err := h.Service.GetViaCep(cep, viaCepUrl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "can not find zipcode"})
		return
	}

	realtimeWeather, err := h.WeatherService.GetWeather(viaCepResponse.Localidade, apiWeatherKey, apiWeatherHost)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "can not find weather"})
		return
	}

	var result dto.Response
	result.TempC = realtimeWeather.TempC
	result.TempF = realtimeWeather.TempF
	result.TempK = realtimeWeather.TempC + 273

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
