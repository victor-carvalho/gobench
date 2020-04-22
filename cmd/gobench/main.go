package main

import (
	"github.com/victor-carvalho/gobench/bencher"
	"github.com/victor-carvalho/gobench/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbosity   = kingpin.Flag("verbose", "Verbosity level (max: 3)").Short('v').Counter()
	concurrency = kingpin.Flag("concurrency", "Number of concurrent connections").Short('c').Default("1")..Int()
	requests    = kingpin.Flag("requests", "Number of requests to make").Short('n').Default("1").Int()
	timeout     = kingpin.Flag("timeout", "Set connection timeout").Short('t').Default("5s").Duration()
	timeLimit   = kingpin.Flag("timelimit", "Maximum number of time to spend for benchmarking").Default("600s").Duration()
	certPath    = kingpin.Flag("cert", "Client certificate file (PEM)").Default("").String()
	keyPath     = kingpin.Flag("key", "Private key file (PEM)").Default("").String()
	serverName  = kingpin.Flag("server", "Server name used to verify the hostname").String()
	status      = kingpin.Flag("status", "Regex to match the response status").Short('s').Default(`2\d\d`).Regexp()
	url         = kingpin.Arg("url", "URL to GET to").Required().URL()
)

func tlsConfig() *config.TLSConfig {
	shouldCreateConfig := false
	tls := new(config.TLSConfig)
	if *certPath != "" && *keyPath != "" {
		tls.KeyPair = &config.KeyPair{CertPath: *certPath, KeyPath: *keyPath}
		shouldCreateConfig = true
	}
	if *serverName != "" {
		tls.ServerName = ""
		shouldCreateConfig = true
	}
	if shouldCreateConfig {
		return tls
	}
	return nil
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("1.0").Author("Victor Carvalho")
	kingpin.CommandLine.Help = "Tool to benchmark HTTP applications in Go"
	kingpin.Parse()

	var cfg config.Config
	cfg.Concurrency = *concurrency
	cfg.MaxRequests = *requests
	cfg.MaxElapsedTime = *timeLimit
	cfg.RequestTimeout = *timeout
	cfg.StatusPattern = *status
	cfg.URL = (*url).String()
	cfg.TLSConfig = tlsConfig()
	cfg.VerbosityLevel = *verbosity

	bch := bencher.NewFromConfig(cfg)
	stats := bch.Run()
	stats.Show()
}
