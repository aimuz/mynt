package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"go.aimuz.me/mynt/store"
)

// policyNameRegex validates policy names: letters, numbers, underscores, hyphens only
var policyNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

func (s *Server) handleListSnapshotPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := s.snapshotPolicy.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, policies)
}

func (s *Server) handleCreateSnapshotPolicy(w http.ResponseWriter, r *http.Request) {
	var policy store.SnapshotPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if policy.Name == "" || policy.Schedule == "" || policy.Retention == "" {
		http.Error(w, "name, schedule, and retention are required", http.StatusBadRequest)
		return
	}

	// Validate policy name format (must be English letters, numbers, underscores, hyphens)
	if !policyNameRegex.MatchString(policy.Name) {
		http.Error(w, "policy name must start with a letter and contain only letters, numbers, underscores, and hyphens", http.StatusBadRequest)
		return
	}

	if err := s.snapshotPolicy.Save(&policy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.notifyPolicyChange()
	respondJSON(w, http.StatusCreated, policy)
}

func (s *Server) handleUpdateSnapshotPolicy(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid policy ID", http.StatusBadRequest)
		return
	}

	// Fetch existing policy first
	existing, err := s.snapshotPolicy.GetByID(id)
	if err != nil {
		http.Error(w, "policy not found", http.StatusNotFound)
		return
	}

	// Decode partial update
	var update struct {
		Name      *string   `json:"name,omitempty"`
		Schedule  *string   `json:"schedule,omitempty"`
		Retention *string   `json:"retention,omitempty"`
		Datasets  *[]string `json:"datasets,omitempty"`
		Enabled   *bool     `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Merge fields
	if update.Name != nil {
		if !policyNameRegex.MatchString(*update.Name) {
			http.Error(w, "policy name must start with a letter and contain only letters, numbers, underscores, and hyphens", http.StatusBadRequest)
			return
		}
		existing.Name = *update.Name
	}
	if update.Schedule != nil {
		existing.Schedule = *update.Schedule
	}
	if update.Retention != nil {
		existing.Retention = *update.Retention
	}
	if update.Datasets != nil {
		existing.Datasets = *update.Datasets
	}
	if update.Enabled != nil {
		existing.Enabled = *update.Enabled
	}

	if err := s.snapshotPolicy.Update(existing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.notifyPolicyChange()
	respondJSON(w, http.StatusOK, existing)
}

func (s *Server) handleDeleteSnapshotPolicy(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid policy ID", http.StatusBadRequest)
		return
	}

	if err := s.snapshotPolicy.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.notifyPolicyChange()
	w.WriteHeader(http.StatusNoContent)
}

// notifyPolicyChange calls the onPolicyChange callback if set.
func (s *Server) notifyPolicyChange() {
	if s.onPolicyChange != nil {
		s.onPolicyChange()
	}
}
