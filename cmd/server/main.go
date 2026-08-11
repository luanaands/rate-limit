package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luanaands/rate-limit/configs"
	_ "github.com/luanaands/rate-limit/docs"
	"github.com/luanaands/rate-limit/internal/infra/service"
	"github.com/luanaands/rate-limit/internal/infra/webserver/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Desafio CEP API - golang
// @version 1.0
// @description API para consulta do tempo real de um CEP utilizando a API do ViaCEP e da WeatherAPI.
// @termsOfService http://swagger.io/terms/

// @contact.name Luana Andrade
// @contact.email luanaands@gmail.com

// @host deploy-cloud-run-1020181349268.us-central1.run.app
// @schemes https
// @basePath /
func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.WithValue("ViaCepHost", configs.ViaCepApiHost))
	r.Use(middleware.WithValue("ApiWeatherHost", configs.ApiWeatherHost))
	r.Use(middleware.WithValue("ApiWeatherKey", configs.ApiWeatherKey))

	var cepService service.CepInterface = service.NewCepService()
	var weatherService service.WeatherInterface = service.NewWeatherService()
	handler := handlers.NewCepHandler(cepService, weatherService)

	r.Get("/weather", handler.GetCep)

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("https://deploy-cloud-run-1020181349268.us-central1.run.app/docs/doc.json")))

	http.ListenAndServe(":8080", r)
}
