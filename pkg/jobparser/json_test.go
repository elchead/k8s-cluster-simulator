package jobparser_test

import (
	"strings"
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
	data := strings.NewReader(`[{"Name": "o10n-worker-l-jbmbp-qsjc6", "Memory": [10553820501, 32589617152, 46905995946, 56151258453, 56794065578, 58443539114],"Time":[1652248860000,1652248865000,1652248865000,1652248865000,1652248865000,1652248869000]}]`)
	jobs, err := jobparser.ParsePodMemoriesFromJson(data)
	assert.NoError(t, err)
	assert.Equal(t, jobs[0].Name, "o10n-worker-l-jbmbp-qsjc6")
	assert.Equal(t, jobs[0].Records[1].Usage, float64(32589617152))
	assert.Equal(t, jobs[0].Records[1].Time, time.Unix(int64(1652248865),0))

	assert.Equal(t, jobs[0].StartAt, time.Unix(int64(1652248860),0))
	assert.Equal(t, jobs[0].EndAt, time.Unix(int64(1652248869),0))
}

// func TestLoadJobs(t *testing.T) {
// 	file, err := os.Open("../../example/pods.json")
// 	assert.NoError(t, err)
// 	jobs,err := jobparser.ParsePodMemoriesFromJson(file)
// 	job := jobparser.FindJob("o10n-worker-l-2xs2w-c7hh4",jobs)
// 	podspec := jobparser.CreatePod(*job)
// 	assert.Equal(t,true,podspec.Annotations["simSpec"])
// }


