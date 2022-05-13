package jobparser_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/stretchr/testify/assert"
)

func TestSubmitJobWhenTime(t *testing.T) {
	now := time.Now()

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
	t.Run("submit job when at exact same time", func(t *testing.T) {
		sut := jobparser.NewJobSubmitter(jobs)
		simTime := clock.NewClock(now)
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(events))
		assertSubmitEvent(t, events[0], "j1")
		assertTerminateEvent(t, events[1])
	})
	t.Run("submit job when before", func(t *testing.T) {
		sut := jobparser.NewJobSubmitter(jobs)
		simTime := clock.NewClock(now.Add(5. * time.Second))
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(events))
		assertSubmitEvent(t, events[0], "j1")
		assertTerminateEvent(t, events[1])
	})
	t.Run("do not submit job when not yet time", func(t *testing.T) {
		sut := jobparser.NewJobSubmitter(jobs)
		simTime := clock.NewClock(now.Add(-5. * time.Second))
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(events))
		// assertSubmitEvent(t, events[0], "j1")
		// assertTerminateEvent(t, events[1])
	})
}

func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	submit, ok := event.(*submitter.SubmitEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, submit.Pod.ObjectMeta.Name)
}

func assertTerminateEvent(t testing.TB, event submitter.Event) {
	_, ok := event.(*submitter.TerminateSubmitterEvent)
	assert.True(t, ok)
}
