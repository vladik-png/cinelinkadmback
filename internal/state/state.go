package state

import (
	"sync"
	"admin-aws/internal/models"
)

var (
	LatestMetrics = make(map[string]models.ServerState)
	MetricsMu     sync.Mutex
)
