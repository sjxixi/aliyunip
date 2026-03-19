package config

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	AccessKeyID     string            `json:"access_key_id"`
	AccessKeySecret string            `json:"access_key_secret"`
	Region          string            `json:"region"`
	Services        map[string]string `json:"services"`
}

const (
	configDirName  = ".aliyun-ip-manager"
	configFileName = "config.json"
)

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configDirName), nil
}

func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

func encodeSecret(secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(secret))
}

func decodeSecret(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func Save(cfg *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	encodedConfig := *cfg
	encodedConfig.AccessKeySecret = encodeSecret(encodedConfig.AccessKeySecret)

	data, err := json.MarshalIndent(&encodedConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	decodedSecret, err := decodeSecret(cfg.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	cfg.AccessKeySecret = decodedSecret

	return &cfg, nil
}

func New() *Config {
	return &Config{
		Services: make(map[string]string),
	}
}
