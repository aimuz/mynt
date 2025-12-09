package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// handleGetSystemStats returns current system statistics.
func (s *Server) handleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	if s.sysMon == nil {
		http.Error(w, "system monitor not available", http.StatusServiceUnavailable)
		return
	}
	respondJSON(w, http.StatusOK, s.sysMon.GetStats())
}

// handleListProcesses returns the list of running processes.
func (s *Server) handleListProcesses(w http.ResponseWriter, r *http.Request) {
	if s.sysMon == nil {
		http.Error(w, "system monitor not available", http.StatusServiceUnavailable)
		return
	}

	procs, err := s.sysMon.GetProcesses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, procs)
}

// handleKillProcess terminates a process by PID.
func (s *Server) handleKillProcess(w http.ResponseWriter, r *http.Request) {
	if s.sysMon == nil {
		http.Error(w, "system monitor not available", http.StatusServiceUnavailable)
		return
	}

	pidStr := r.PathValue("pid")
	pid, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid pid", http.StatusBadRequest)
		return
	}

	if err := s.sysMon.KillProcess(int32(pid)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
