package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
)

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func GetenvStr(key string, default_val string) string {
	v := os.Getenv(key)
	if v == "" {
		logging.Metric("missing_config_value").Warn("missing value for config", zap.String("config_key", key))
		return default_val
	}
	return v
}

func GetenvStrArr(key string, default_val []string) ([]string, error) {
	s := GetenvStr(key, "")
	if s == "" {
		logging.Metric("missing_config_value").Warn("missing value for config", zap.String("config_key", key))
		return default_val, nil
	}
	var arr []string
	err := json.Unmarshal([]byte(s), &arr)
	if err != nil {
		logging.Metric("config_parse_error").Error(
			"failed to parse config, using default",
			zap.String("config_key", key),
			zap.Any("default_arr", default_val),
			zap.Error(err),
		)
		return default_val, err
	}
	return arr, nil
}

func GetenvStrArrErrorless(key string, default_val []string) (res []string) {
	res, _ = GetenvStrArr(key, default_val)
	return
}

func GetenvInt(key string, default_val int) (int, error) {
	s := GetenvStr(key, "")
	if s == "" {
		logging.Metric("missing_config_value").Warn("missing value for config", zap.String("config_key", key))
		return default_val, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		logging.Metric("config_parse_error").Error(
			"failed to parse config, using default",
			zap.String("config_key", key),
			zap.Int("default_int", default_val),
			zap.Error(err),
		)
		return default_val, err
	}
	return v, nil
}

func GetenvIntErrorless(key string, default_val int) (res int) {
	res, _ = GetenvInt(key, default_val)
	return
}

func GetenvBool(key string, default_val bool) (bool, error) {
	s := GetenvStr(key, "")
	if s == "" {
		logging.Metric("missing_config_value").Warn("missing value for config", zap.String("config_key", key))
		return default_val, nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		logging.Metric("config_parse_error").Error(
			"failed to parse config, using default",
			zap.String("config_key", key),
			zap.Bool("default_bool", default_val),
			zap.Error(err),
		)
		return default_val, err
	}
	return v, nil
}

func GetenvBoolErrorless(key string, default_val bool) (res bool) {
	res, _ = GetenvBool(key, default_val)
	return
}
