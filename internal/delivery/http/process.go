package httpdelivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type KillProcessRequest struct {
	InstanceID string `json:"instance_id"`
	PID        int    `json:"pid"`
}

func HandleKillProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req KillProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}


	var err error
	if runtime.GOOS == "windows" {
		err = exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", req.PID)).Run()
	} else {
		err = exec.Command("kill", "-9", fmt.Sprintf("%d", req.PID)).Run()
	}

	if err != nil {
		p, e := os.FindProcess(req.PID)
		if e == nil {
			err = p.Kill()
		}
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Failed to kill process: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
	})
}
