package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/luanaands/rate-limit/internal/dto"
	"github.com/luanaands/rate-limit/internal/entity"
)

type CepService struct {
	client *http.Client
}

func NewCepService() *CepService {
	return &CepService{
		client: &http.Client{},
	}
}

func (s *CepService) GetViaCep(cep string, url string) (*dto.CepResponse, error) {
	req, err := http.NewRequest("GET", url+"/"+cep+"/json", nil)
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
	var response *entity.CepViaCepResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Erro == "true" {
		return nil, errors.New("can not find zipcode")
	}

	var dtoResponse *dto.CepResponse
	dtoResponse = dto.FromViaCep(response)
	return dtoResponse, nil
}
