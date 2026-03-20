package config

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Config struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region"`
}

type HistoryRecord struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	IPAddress   string            `json:"ip_address"`
	Port        int               `json:"port"`
	Description string            `json:"description"`
	Resources   []HistoryResource `json:"resources"`
}

type HistoryResource struct {
	Type            string `json:"type"`
	Id              string `json:"id"`
	Name            string `json:"name"`
	SecurityIpGroup string `json:"security_ip_group,omitempty"`
	Port            string `json:"port,omitempty"`
	Description     string `json:"description,omitempty"`
}

const (
	configDirName     = ".aliyun-ip-manager"
	configFileName    = "config.json"
	historyFileName   = "history.json"
	maxHistoryRecords = 5
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
	return &Config{}
}

func GetHistoryPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, historyFileName), nil
}

func SaveHistory(record HistoryRecord) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	historyPath, err := GetHistoryPath()
	if err != nil {
		return err
	}

	var history []HistoryRecord

	if data, err := os.ReadFile(historyPath); err == nil {
		if err := json.Unmarshal(data, &history); err != nil {
			history = []HistoryRecord{}
		}
	}

	history = append([]HistoryRecord{record}, history...)

	if len(history) > maxHistoryRecords {
		history = history[:maxHistoryRecords]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0600)
}

func LoadHistory() ([]HistoryRecord, error) {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryRecord{}, nil
		}
		return nil, err
	}

	var history []HistoryRecord
	if err := json.Unmarshal(data, &history); err != nil {
		return []HistoryRecord{}, nil
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.After(history[j].Timestamp)
	})

	return history, nil
}
