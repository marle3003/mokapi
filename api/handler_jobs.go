package api

import (
	"mokapi/engine/common"
	"net/http"
	"time"
)

type jobInfo struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Schedule string            `json:"schedule"`
	MaxRuns  int               `json:"maxRuns"`
	Runs     int               `json:"runs"`
	State    string            `json:"state"`
	NextRun  *time.Time        `json:"nextRun,omitempty"`
	Tags     map[string]string `json:"tags"`
	LastRun  *common.JobRun    `json:"lastRun,omitempty"`
}

func (h *handler) getJobs(w http.ResponseWriter, _ *http.Request) {
	jobs := make([]jobInfo, 0)
	for _, j := range h.app.Jobs() {
		jobs = append(jobs, toJobInfo(j))
	}
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, jobs)
}

func toJobInfo(j common.JobInfo) jobInfo {
	info := jobInfo{
		Id:       j.Id,
		Name:     j.Name,
		Schedule: j.Schedule,
		MaxRuns:  j.MaxRuns,
		Runs:     j.Runs,
		State:    j.State,
		Tags:     j.Tags,
		LastRun:  j.LastRun,
	}
	if !j.NextRun.IsZero() {
		info.NextRun = &j.NextRun
	}
	return info
}
