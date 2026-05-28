package httpdelivery

import (
	"encoding/json"
	"net/http"

	"admin-aws/internal/database"
	"admin-aws/internal/models"
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
	}
	w.WriteHeader(http.StatusOK)
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	var logs []models.ServerLog
	if database.DB != nil {
		database.DB.Order("created_at desc").Limit(50).Find(&logs)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	var alerts []models.ServerAlert
	if database.DB != nil {
		database.DB.Where("resolved = ?", false).Order("created_at desc").Find(&alerts)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}
