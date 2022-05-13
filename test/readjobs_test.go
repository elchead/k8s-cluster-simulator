package test

import (
	"os"
	"testing"

	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/stretchr/testify/assert"
)

func TestCreatePodsFromJobCSV(t *testing.T) {
	f, err := os.Open("./pods11-05-8to12.csv")
	assert.NoError(t, err)
	podmems := jobparser.ParsePodMemories(f)
	for _, pod := range podmems {
		spec := jobparser.CreatePod(pod)
		assert.NotEmpty(t, spec.ObjectMeta.Name)

	}
	assert.NotEmpty(t, podmems)
}
