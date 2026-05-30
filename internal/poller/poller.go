package poller

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Poller struct {
	client   *http.Client
	url      string
	interval time.Duration
}

func NewPoller(url string, interval time.Duration) *Poller {
	return &Poller{
		url:      url,
		interval: interval,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				MaxConnsPerHost:     50,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	log.Printf("poller started: url=%s interval=%s", p.url, p.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("poller stopped")
			return

		case <-ticker.C:
			p.callDownStream(ctx)
		}
	}
}

func (p *Poller) callDownStream(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return
	}

	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("downstream request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("downstream response status=%d", resp.StatusCode)
}
