package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/spf13/viper"
)

func Load() (*Configs, error) {
	configuration := newConfig()

	var appConfig Configs
	err := configuration.Unmarshal(&appConfig)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Failed to load configuration",
			ServiceMessage:    fmt.Sprintf("Error unmarshaling config: %v", err),
		}
	}

	// Validate the configuration
	if err := validator.Validate.Struct(&appConfig); err != nil {
		return nil, toddlerr.FromValidationErrors(err)
	}

	return &appConfig, nil
}

func newConfig() *viper.Viper {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath(".")
	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("got an error reading file config, error: %s", err)
	}

	// Replace placeholders like ${VAR_NAME} with environment variables after reading the config
	replacePlaceholdersWithEnv(config)

	return config
}

// This function will iterate through all keys in the config and replace any placeholders like ${VAR_NAME} with their environment values
func replacePlaceholdersWithEnv(config *viper.Viper) {
	for _, key := range config.AllKeys() {
		switch v := config.Get(key).(type) {
		case string:
			if sub, ok := substituteEnv(v); ok {
				config.Set(key, sub)
			}
		case []interface{}:
			out := make([]interface{}, len(v))
			changed := false
			for i, item := range v {
				s, ok := item.(string)
				if !ok {
					out[i] = item
					continue
				}
				if sub, didSub := substituteEnv(s); didSub {
					out[i] = sub
					changed = true
					continue
				}
				out[i] = s
			}
			if changed {
				config.Set(key, out)
			}
		}
	}
}

func substituteEnv(value string) (string, bool) {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value, false
	}
	envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
	envValue := os.Getenv(envVar)
	if envValue == "" {
		log.Fatalf("Mandatory environment variable %s not found", envVar)
	}
	return envValue, true
}
