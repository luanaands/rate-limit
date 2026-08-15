package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/luanaands/rate-limit/configs"
	_ "github.com/luanaands/rate-limit/docs"
	"github.com/luanaands/rate-limit/internal/infra/service"
	"github.com/luanaands/rate-limit/internal/infra/webserver/handlers"
	rateLimitMiddleware "github.com/luanaands/rate-limit/internal/infra/webserver/middleware"
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
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		panic(fmt.Errorf("connect to redis: %w", err))
	}

	rateLimiter := rateLimitMiddleware.NewRateLimiterConfig(
		cfg.RateLimitTokenLimit,
		cfg.RateLimitIPLimit,
		cfg.RateLimitBlockDuration,
		redisClient,
	)
	rateLimitProcessor := rateLimitMiddleware.NewRateLimitProcessor(rateLimiter)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			allowed, _ := rateLimitProcessor.Allow(request.Context(), request)
			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(cfg.RateLimitBlockDuration.Seconds())))
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, request)
		})
	})
	r.Use(middleware.WithValue("ViaCepHost", cfg.ViaCepApiHost))
	r.Use(middleware.WithValue("ApiWeatherHost", cfg.ApiWeatherHost))
	r.Use(middleware.WithValue("ApiWeatherKey", cfg.ApiWeatherKey))

	var cepService service.CepInterface = service.NewCepService()
	var weatherService service.WeatherInterface = service.NewWeatherService()
	handler := handlers.NewCepHandler(cepService, weatherService)

	r.Get("/weather", handler.GetCep)

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("doc.json")))

	http.ListenAndServe(":8080", r)
}
