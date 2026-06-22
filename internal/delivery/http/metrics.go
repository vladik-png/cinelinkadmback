package httpdelivery

import (
	"encoding/json"
	"net/http"
	"time"

	"admin-aws/internal/database"
	"admin-aws/internal/models"
	"admin-aws/internal/state"
)

func ReceiveMetricsFromAgent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, ok := data["instance_id"].(string)
	if !ok {
		return
	}

	state.MetricsMu.Lock()
	state.LatestMetrics[id] = models.ServerState{
		LastSeen: time.Now(),
		Metrics:  data,
	}

	frontData := make(map[string]interface{})
	for k, s := range state.LatestMetrics {
		frontData[k] = s.Metrics
	}
	state.MetricsMu.Unlock()

	LiveHub.Broadcast("metrics", frontData)

	if database.DB != nil {
		database.DB.Model(&models.ServerAlert{}).Where("server_id = ? AND type = ? AND resolved = ?", id, "OFFLINE", false).Update("resolved", true)
	}
	w.WriteHeader(http.StatusOK)
}

func GetMetricsForFront(w http.ResponseWriter, r *http.Request) {
	state.MetricsMu.Lock()
	defer state.MetricsMu.Unlock()

	frontData := make(map[string]interface{})
	for id, s := range state.LatestMetrics {
		frontData[id] = s.Metrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(frontData)
}
