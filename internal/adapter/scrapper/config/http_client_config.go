package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

var (
	ErrNoTimeout       = errors.New("http client timeout should be set with" + clientTimeout + " environment variable")
	ErrInvalidTimeout  = errors.New("http client timeout " + clientTimeout + "should be integer")
	ErrNegativeTimeout = errors.New("http client timeout " + clientTimeout + "should be positive value")
)

const (
	clientTimeout = "HTTP_CLIENT_TIMEOUT_SEC"
)

type HTTPClientConfig struct {
	Timeout time.Duration
}

func ParseHTTPClientConfig() (HTTPClientConfig, error) {
	timeoutStr := os.Getenv(clientTimeout)
	if timeoutStr == "" {
		return HTTPClientConfig{}, ErrNoTimeout
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return HTTPClientConfig{}, ErrInvalidTimeout
	}
	if timeout < 0 {
		return HTTPClientConfig{}, ErrNegativeTimeout
	}
	return HTTPClientConfig{Timeout: time.Duration(timeout) * time.Second}, nil
}
