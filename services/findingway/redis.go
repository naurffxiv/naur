package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

const redisAddr = "redis.hyddwn.net:6380"

// newRedisClient builds the client for the description log from
// REDIS_USER, REDIS_PASSWORD and the mTLS certificate files.
func newRedisClient() (*redis.Client, error) {
	redisPw, ok := os.LookupEnv("REDIS_PASSWORD")
	redisUser, userOk := os.LookupEnv("REDIS_USER")

	if !userOk {
		return nil, errors.New("REDIS_USER is not set")
	}
	if !ok {
		return nil, errors.New("REDIS_PASSWORD is not set")
	}

	cert, err := tls.LoadX509KeyPair("naur.crt", "naur.key")
	if err != nil {
		return nil, fmt.Errorf("error loading certificates: %w", err)
	}

	caCert, err := os.ReadFile("hyddwn-ca.crt")
	if err != nil {
		return nil, fmt.Errorf("error loading CA certificate: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append CA certificate")
	}

	return redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: redisUser,
		Password: redisPw,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
		},
		// Honour each call's context deadline so a hung Redis cannot stall
		// the listing loop.
		ContextTimeoutEnabled: true,
	}), nil
}
