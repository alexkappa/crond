package main

import (
	"net/http"
	"regexp"
	"time"
)

type healthCheck struct {
	regexp *regexp.Regexp
	client *http.Client
}

func (h *healthCheck) HasAnnotation(s string) bool {
	return h.regexp.MatchString(s)
}

func (h *healthCheck) FromAnnotation(s string) string {
	return h.regexp.FindString(s)[5:]
}

func (h *healthCheck) Ping(check string) error {
	r, err := http.NewRequest("GET", "https://hchk.io/"+check, nil)
	if err != nil {
		return err
	}
	r.Header.Set("User-Agent", "Cron/"+Version)
	_, err = h.client.Do(r)
	return err
}

func newHealthCheck() *healthCheck {
	return &healthCheck{
		regexp: regexp.MustCompile(`# hc:[[:alnum:]]{8}-([[:alnum:]]{4}-){3}[[:alnum:]]{12}`),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
