package fetchup

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Fetchup struct {
	To   string
	URLs []string

	Logger          Logger
	IdleConnTimeout time.Duration
	MinReportSpan   time.Duration
	HttpClient      *http.Client
}

func New(to string, us ...string) *Fetchup {
	return &Fetchup{
		To:              to,
		URLs:            us,
		Logger:          log.Default(),
		IdleConnTimeout: 30 * time.Second,
		MinReportSpan:   time.Second,
		HttpClient:      http.DefaultClient,
	}
}

func (fu *Fetchup) Fetch() error {
	u := fu.FastestURL()
	return fu.Download(u)
}

func (fu *Fetchup) FastestURL() (fastest string) {
	setURL := sync.Once{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	for _, u := range fu.URLs {
		u := u

		wg.Add(1)

		go func() {
			defer wg.Done()

			q, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return
			}

			res, err := fu.HttpClient.Do(q)
			if err != nil {
				return
			}
			defer func() { _ = res.Body.Close() }()

			if res.StatusCode == http.StatusOK {
				buf := make([]byte, 64*1024) // a TCP packet won't be larger than 64KB
				_, err = res.Body.Read(buf)
				if err != nil {
					return
				}

				setURL.Do(func() {
					fastest = u
					cancel()
				})
			}
		}()
	}
	wg.Wait()

	return
}
