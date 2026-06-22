package httpdelivery

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"admin-aws/internal/database"
	"admin-aws/internal/models"
)

var (
	logsCache     []models.ServerLog
	logsCacheTime time.Time
	logsMutex     sync.Mutex

	alertsCache     []models.ServerAlert
	alertsCacheTime time.Time
	alertsMutex     sync.Mutex
)

func HandleLogEvent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	instanceID, _ := data["instance_id"].(string)
	action, _ := data["action"].(string)
	status, _ := data["status"].(string)
	details, _ := data["details"].(string)
	
	if instanceID != "" {
		database.LogEvent(instanceID, action, status, details)

		logsMutex.Lock()
		logsCacheTime = time.Time{}
		logsMutex.Unlock()

		alertsMutex.Lock()
		alertsCacheTime = time.Time{}
		alertsMutex.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	logsMutex.Lock()
	if time.Since(logsCacheTime) < 5*time.Second && logsCache != nil {
		logs := logsCache
		logsMutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
		return
	}
	logsMutex.Unlock()

	var logs []models.ServerLog
	if database.DB != nil {
		database.DB.Order("created_at desc").Limit(50).Find(&logs)
	}

	logsMutex.Lock()
	logsCache = logs
	logsCacheTime = time.Now()
	logsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	alertsMutex.Lock()
	if time.Since(alertsCacheTime) < 5*time.Second && alertsCache != nil {
		alerts := alertsCache
		alertsMutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
		return
	}
	alertsMutex.Unlock()

	var alerts []models.ServerAlert
	if database.DB != nil {
		database.DB.Where("resolved = ?", false).Order("created_at desc").Find(&alerts)
	}

	alertsMutex.Lock()
	alertsCache = alerts
	alertsCacheTime = time.Now()
	alertsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}
