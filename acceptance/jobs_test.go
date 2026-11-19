package acceptance

import (
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

type jobsApiJob struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Schedule string            `json:"schedule"`
	MaxRuns  int               `json:"maxRuns"`
	Runs     int               `json:"runs"`
	State    string            `json:"state"`
	NextRun  *time.Time        `json:"nextRun"`
	Tags     map[string]string `json:"tags"`
	LastRun  *struct {
		EventId  string    `json:"eventId"`
		Time     time.Time `json:"time"`
		Duration int64     `json:"duration"`
		Status   string    `json:"status"`
	} `json:"lastRun"`
}

type JobsSuite struct{ BaseSuite }

func (suite *JobsSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./jobs"}}
	suite.initCmd(cfg)
}

func (suite *JobsSuite) apiUrl(path string) string {
	return fmt.Sprintf("http://127.0.0.1:%v%v", suite.cfg.Api.Port, path)
}

func (suite *JobsSuite) getJobs() []jobsApiJob {
	res, err := http.Get(suite.apiUrl("/api/jobs"))
	suite.Require().NoError(err)
	defer res.Body.Close()
	suite.Require().Equal(http.StatusOK, res.StatusCode)
	var jobs []jobsApiJob
	suite.Require().NoError(json.NewDecoder(res.Body).Decode(&jobs))
	return jobs
}

// waitForJob waits until the jobs of the script file are listed and the first job satisfies
// the condition. Scripts are loaded asynchronously and jobs run in the background.
func (suite *JobsSuite) waitForJob(file string, condition func(j jobsApiJob) bool) jobsApiJob {
	suite.T().Helper()
	var found jobsApiJob
	try.WaitFor(suite.T(), func() bool {
		res, err := http.Get(suite.apiUrl("/api/jobs"))
		if err != nil {
			return false
		}
		defer res.Body.Close()
		suite.Require().Equal(http.StatusOK, res.StatusCode)
		var jobs []jobsApiJob
		if json.NewDecoder(res.Body).Decode(&jobs) != nil {
			return false
		}
		for _, j := range jobs {
			if filepath.Base(j.Tags["file"]) == file && condition(j) {
				found = j
				return true
			}
		}
		return false
	})
	return found
}

func (suite *JobsSuite) jobsOfFile(file string) []jobsApiJob {
	var result []jobsApiJob
	for _, j := range suite.getJobs() {
		if filepath.Base(j.Tags["file"]) == file {
			result = append(result, j)
		}
	}
	return result
}

func (suite *JobsSuite) TestEveryJobIsListedWithScheduleRunsAndNextRun() {
	j := suite.waitForJob("every.js", func(j jobsApiJob) bool { return j.Runs >= 1 })

	suite.Equal("every.js", j.Name)
	suite.Equal("5m", j.Schedule)
	suite.Equal("scheduled", j.State)
	suite.NotNil(j.NextRun)
	suite.True(j.NextRun.After(time.Now()))
	suite.NotEmpty(j.Id)
	suite.NotEmpty(j.Tags["fileKey"])
}

func (suite *JobsSuite) TestJobThatHasNotRunYetIsListedWithoutLastRun() {
	j := suite.waitForJob("later.js", func(j jobsApiJob) bool { return true })

	suite.Equal("later", j.Name)
	suite.Equal(0, j.Runs)
	suite.Nil(j.LastRun)
	suite.Equal("scheduled", j.State)
	suite.NotNil(j.NextRun)
}

func (suite *JobsSuite) TestLimitedEveryJobFinishesWithoutNextRun() {
	j := suite.waitForJob("limited.js", func(j jobsApiJob) bool { return j.State == "finished" })

	suite.Equal(2, j.MaxRuns)
	suite.Equal(2, j.Runs)
	suite.Nil(j.NextRun)
	suite.NotNil(j.LastRun)
}

// The immediate first run counts as one of the times runs, so the job is finished right away.
func (suite *JobsSuite) TestCronJobWithTimesFinishesWithImmediateFirstRun() {
	j := suite.waitForJob("cron.js", func(j jobsApiJob) bool { return j.MaxRuns == 1 && j.Runs >= 1 })

	suite.Equal("* * * * *", j.Schedule)
	suite.Equal(1, j.Runs)
	suite.Equal("finished", j.State)
	suite.Nil(j.NextRun)
}

func (suite *JobsSuite) TestLastRunLinksToTheEventOfTheRun() {
	tests := []struct {
		file   string
		status string
	}{
		{file: "every.js", status: "success"},
		{file: "warning.js", status: "warning"},
		{file: "failing.js", status: "error"},
		{file: "throws.js", status: "error"},
	}
	for _, tc := range tests {
		suite.T().Run(tc.file, func(t *testing.T) {
			j := suite.waitForJob(tc.file, func(j jobsApiJob) bool { return j.LastRun != nil })

			suite.Equal(tc.status, j.LastRun.Status)
			suite.Equal(1, j.Runs, "a failed run is counted too")
			try.GetRequest(t, suite.apiUrl("/api/events/"+j.LastRun.EventId), nil,
				try.HasStatusCode(http.StatusOK),
				try.BodyContains(`"id":"`+j.LastRun.EventId+`"`),
			)
		})
	}
}

func (suite *JobsSuite) TestJobsAreSortedByNameThenFile() {
	expected := []string{
		"cron.js cron.js",
		"cron.js cron.js",
		"every.js every.js",
		"failing.js failing.js",
		"later later.js",
		"limited.js limited.js",
		"shared shared-a.js",
		"shared shared-b.js",
		"throws.js throws.js",
		"warning.js warning.js",
	}

	// scripts run in parallel, so wait until every script has registered its jobs
	var jobs []jobsApiJob
	try.WaitFor(suite.T(), func() bool {
		jobs = suite.getJobs()
		return len(jobs) == len(expected)
	})

	var order []string
	for _, j := range jobs {
		order = append(order, j.Name+" "+filepath.Base(j.Tags["file"]))
	}
	suite.Equal(expected, order)
}
