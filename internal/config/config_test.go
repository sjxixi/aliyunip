package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigOperations(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	testDir := filepath.Join(os.TempDir(), "aliyunip_test")
	if err := os.MkdirAll(testDir, 0700); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	os.Setenv("HOME", testDir)

	cfg := New()
	cfg.AccessKeyID = "test-access-key-id"
	cfg.AccessKeySecret = "test-access-key-secret"
	cfg.Region = "cn-shanghai-finance-1"

	if err := Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.AccessKeyID != cfg.AccessKeyID {
		t.Errorf("AccessKeyID mismatch: expected %s, got %s", cfg.AccessKeyID, loadedCfg.AccessKeyID)
	}

	if loadedCfg.AccessKeySecret != cfg.AccessKeySecret {
		t.Errorf("AccessKeySecret mismatch: expected %s, got %s", cfg.AccessKeySecret, loadedCfg.AccessKeySecret)
	}

	if loadedCfg.Region != cfg.Region {
		t.Errorf("Region mismatch: expected %s, got %s", cfg.Region, loadedCfg.Region)
	}
}
