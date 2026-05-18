package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoValkeyAddresses = errors.New("valkey addresses should be set with " + valkeyAddressesEnv + " environment variable")
	ErrNoValkeyPassword  = errors.New("valkey password should be set with " + valkeyPasswordEnv + " environment variable")
	ErrNoValkeyTTL       = errors.New("valkey ttl should be set with " + valkeyTTLEnv + " environment variable")
	ErrInvalidValkeyTTL  = errors.New("valkey ttl should be integer value")
)

const (
	valkeyAddressesEnv = "VALKEY_ADDRESSES"
	valkeyPasswordEnv  = "VALKEY_PASSWORD"
	valkeyTTLEnv       = "VALKEY_TTL_MINUTES"
)

type ValkeyConfig struct {
	Addresses []string
	Password  string
	ValkeyTTL time.Duration
}

func ParseValkeyConfig() (ValkeyConfig, error) {
	valkeyAddresses := os.Getenv(valkeyAddressesEnv)
	if valkeyAddresses == "" {
		return ValkeyConfig{}, ErrNoValkeyAddresses
	}
	valkeyPassword := os.Getenv(valkeyPasswordEnv)
	if valkeyPassword == "" {
		return ValkeyConfig{}, ErrNoValkeyPassword
	}
	valkeyTTLStr := os.Getenv(valkeyTTLEnv)
	if valkeyTTLStr == "" {
		return ValkeyConfig{}, ErrNoValkeyTTL
	}
	valkeyTTL, err := strconv.Atoi(valkeyTTLStr)
	if err != nil {
		return ValkeyConfig{}, ErrInvalidValkeyTTL
	}
	return ValkeyConfig{Addresses: strings.Split(valkeyAddresses, ","), Password: valkeyPassword, ValkeyTTL: time.Duration(valkeyTTL) * time.Minute}, nil
}
