package config

import (
	"regexp"
	"time"
)

type KeyPair struct {
	CertPath string
	KeyPath  string
}

type TLSConfig struct {
	KeyPair    *KeyPair
	ServerName string
}

type Config struct {
	VerbosityLevel int
	Concurrency    int
	MaxRequests    int
	RequestTimeout time.Duration
	MaxElapsedTime time.Duration
	URL            string
	TLSConfig      *TLSConfig
	StatusPattern  *regexp.Regexp
}
