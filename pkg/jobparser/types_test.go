package jobparser_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/stretchr/testify/assert"
)


func TestGetJob(t *testing.T){
	now := time.Now()
	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}

	t.Run("Get job", func(t *testing.T){
		job := jobparser.GetJob("j2",jobs)
		assert.NotNil(t, job)
		assert.Equal(t,"j2",job.Name)
	})
	t.Run("Changing the job pointer also mutates the original slice",func(t *testing.T){
		job := jobparser.GetJob("j2",jobs)
		job.Name = "mj2"
		assert.Equal(t,"mj2",jobs[1].Name)
	})
}
