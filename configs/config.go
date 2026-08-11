package configs

import (
	"github.com/spf13/viper"
)

type Conf struct {
	ViaCepApiHost  string `mapstructure:"VIA_CEP_API_HOST"`
	ApiWeatherKey  string `mapstructure:"API_WEATHER_KEY"`
	ApiWeatherHost string `mapstructure:"API_WEATHER_HOST"`
}

func LoadConfig(path string) (*Conf, error) {
	var cfg *Conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg, err
}
