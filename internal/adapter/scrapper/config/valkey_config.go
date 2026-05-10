package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrNoValkeyAddresses = errors.New("valkey addresses should be set with " + valkeyAddressesEnv + " environment variable")
	ErrNoValkeyPassword  = errors.New("valkey password should be set with " + valkeyPasswordEnv + " environment variable")
)

const (
	valkeyAddressesEnv = "VALKEY_ADDRESSES"
	valkeyPasswordEnv  = "VALKEY_PASSWORD"
)

type ValkeyConfig struct {
	Addresses []string
	Password  string
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
	return ValkeyConfig{Addresses: strings.Split(valkeyAddresses, ","), Password: valkeyPassword}, nil
}
