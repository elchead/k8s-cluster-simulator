package jobparser_test

import (
	"testing"

	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
	data := `[{"Name": "o10n-worker-l-jbmbp-qsjc6", "Memory": [10553820501, 32589617152, 46905995946, 56151258453, 56794065578, 58443539114]}]`
	jobs, err := jobparser.ParseJson([]byte(data))
	assert.NoError(t, err)
	assert.Equal(t, jobs[0].Name, "o10n-worker-l-jbmbp-qsjc6")
	assert.Equal(t, jobs[0].Memory[1], int64(32589617152))
}
