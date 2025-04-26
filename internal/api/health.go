package api

import "context"

type Health struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Time    string `json:"time"`
	Details any    `json:"details,omitempty"`
}

type HealthResponse struct {
	Health []Health `json:"health"`
}

type HealthChecker interface {
	CheckHealth(ctx context.Context) ([]Health, error)
}

type HealthCheckerFunc func(ctx context.Context) ([]Health, error)

func (h HealthCheckerFunc) CheckHealth(ctx context.Context) ([]Health, error) { return h(ctx) }
