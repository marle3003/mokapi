package engine

import (
	"fmt"
	"maps"
	"mokapi/engine/common"
	"sort"
	"sync"
)

// jobEntry is a job registered by a script: the scheduler's job, what the script defined and the
// state that changes with every run.
type jobEntry struct {
	job      Job
	schedule string
	maxRuns  int
	tags     map[string]string
	state    *jobState
}

// jobState is shared between the job's runs (scheduler goroutines) and the readers of the job
// list. Its mutex is a leaf lock: nothing else is acquired while holding it.
type jobState struct {
	maxRuns int

	mu        sync.Mutex
	runs      int
	lastRun   *common.JobRun
	cancelled bool
}

type jobSnapshot struct {
	runs    int
	lastRun *common.JobRun
	state   string
}

func newJobState(maxRuns int) *jobState {
	return &jobState{maxRuns: maxRuns}
}

// nextRunNumber is the number of the run that is about to finish.
func (s *jobState) nextRunNumber() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.runs + 1
}

// finishRun counts the run, also a failed one.
func (s *jobState) finishRun(run common.JobRun) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs++
	s.lastRun = &run
}

func (s *jobState) cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelled = true
}

func (s *jobState) snapshot() jobSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := common.JobStateScheduled
	switch {
	case s.cancelled:
		state = common.JobStateCancelled
	case s.maxRuns > 0 && s.runs >= s.maxRuns:
		state = common.JobStateFinished
	}
	return jobSnapshot{runs: s.runs, lastRun: s.lastRun, state: state}
}

// listJobs describes the registered jobs ordered by their index.
func (sh *scriptHost) listJobs() []common.JobInfo {
	entries := sh.copyJobs()

	ids := make([]int, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	infos := make([]common.JobInfo, 0, len(ids))
	for _, id := range ids {
		infos = append(infos, sh.toJobInfo(id, entries[id]))
	}
	return infos
}

func (sh *scriptHost) copyJobs() map[int]*jobEntry {
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	return maps.Clone(sh.jobs)
}

func (sh *scriptHost) toJobInfo(id int, entry *jobEntry) common.JobInfo {
	snapshot := entry.state.snapshot()
	info := common.JobInfo{
		// not tags["fileKey"], which scripts can override
		Id:       fmt.Sprintf("%s-%d", sh.file.Info.Key(), id),
		Name:     entry.tags["name"],
		Schedule: entry.schedule,
		MaxRuns:  entry.maxRuns,
		Runs:     snapshot.runs,
		State:    snapshot.state,
		Tags:     maps.Clone(entry.tags),
		LastRun:  snapshot.lastRun,
	}
	if snapshot.state == common.JobStateScheduled && entry.job != nil {
		info.NextRun = entry.job.NextRun()
	}
	return info
}
