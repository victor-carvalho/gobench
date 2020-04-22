package requester

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/tcnksm/go-httpstat"
	"github.com/victor-carvalho/gobench/config"
)

func worker(ctx context.Context, wg *sync.WaitGroup, url string, client *http.Client, reqCh chan Request, outCh chan Response) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-reqCh:
			if !ok {
				break
			}
			var response Response

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Fatal("Error creating request")
			}

			ctx := httpstat.WithHTTPStat(req.Context(), &response.Stat)
			req = req.WithContext(ctx)

			resp, err := client.Do(req)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					response.Timeout = true
				} else {
					response.Timeout = false
				}
				response.Error = true
			} else {
				response.StatusCode = resp.StatusCode
				response.Status = resp.Status
			}
			outCh <- response
		case <-ctx.Done():
			return
		}
	}
}

type Requester struct {
	client      *http.Client
	concurrency int
	maxRequests int
	url         string
	requestCh   chan Request
	outputCh    chan Response
}

func NewRequester(cfg config.Config) *Requester {
	client := new(http.Client)
	client.Timeout = cfg.RequestTimeout
	transport := new(http.Transport)
	if cfg.TLSConfig != nil {
		tlsConfig := new(tls.Config)

		if cfg.TLSConfig.KeyPair != nil {
			cert, err := tls.LoadX509KeyPair(cfg.TLSConfig.KeyPair.CertPath, cfg.TLSConfig.KeyPair.KeyPath)
			if err != nil {
				log.Fatal(err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if cfg.TLSConfig.ServerName != "" {
			tlsConfig.ServerName = cfg.TLSConfig.ServerName
		}

		transport.TLSClientConfig = tlsConfig
	}
	client.Transport = transport
	return &Requester{
		url:         cfg.URL,
		client:      client,
		concurrency: cfg.Concurrency,
		maxRequests: cfg.MaxRequests,
		requestCh:   make(chan Request, cfg.MaxRequests),
		outputCh:    make(chan Response, cfg.MaxRequests),
	}
}

func (r *Requester) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(r.concurrency)
	for i := 0; i < r.concurrency; i++ {
		go worker(ctx, &wg, r.url, r.client, r.requestCh, r.outputCh)
	}
	for i := 0; i < r.maxRequests; i++ {
		r.requestCh <- Request{i}
	}
	wg.Wait()
	close(r.requestCh)
	close(r.outputCh)
}

func (r *Requester) Output() chan Response {
	return r.outputCh
}
