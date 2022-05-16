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

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}, {Time: now.Add(1 * time.Hour), Usage: 100.}}}}
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
		assert.Empty(t, events)
	})
	t.Run("remove jobs from submitter once submitted", func(t *testing.T) {
		jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
		sut := jobparser.NewJobSubmitter(jobs)
		simTime := clock.NewClock(now)
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(events))
		assertSubmitEvent(t, events[0], "j1")
		assertSubmitEvent(t, events[1], "j2")
		assertTerminateEvent(t, events[2])

		events, err = sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assertTerminateEvent(t, events[0])
	})
}

func TestIterator(t *testing.T) {
	now := time.Now()
	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}, {Time: now.Add(1 * time.Hour), Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}, {Time: now.Add(1 * time.Hour), Usage: 100.}}}}
	sut := jobparser.NewIterator(jobs)
	val := sut.Value()
	assert.Equal(t, jobs[0], val)
	assert.Equal(t, 2, sut.RemainingValues())

	assert.True(t, sut.Next())
	assert.Equal(t, jobs[1], sut.Value())
	assert.Equal(t, 1, sut.RemainingValues())

	assert.False(t, sut.Next())
	assert.Equal(t, 0, sut.RemainingValues())
	t.Run("next for len(1)", func(t *testing.T) {
		jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
		sut := jobparser.NewIterator(jobs)
		assert.False(t, sut.Next())

	})
	t.Run("next for len(2)", func(t *testing.T) {
		jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
		sut := jobparser.NewIterator(jobs)
		assert.True(t, sut.Next())
		assert.False(t, sut.Next())

	})
	t.Run("empty iterator", func(t *testing.T) {
		jobs := []jobparser.PodMemory{}
		sut := jobparser.NewIterator(jobs)
		assert.False(t, sut.Next())
		job := jobparser.PodMemory{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}
		sut.Push(job)
		assert.Equal(t,job, sut.Value())
		assert.Equal(t,1, sut.RemainingValues())

	})
}

func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	submit, ok := event.(*submitter.SubmitEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, submit.Pod.ObjectMeta.Name)
}

func assertTerminateEvent(t testing.TB, event submitter.Event) {
	assert.True(t, isTerminateEvent(event))
}

func isTerminateEvent(event submitter.Event) (ok bool) {
	_, ok = event.(*submitter.TerminateSubmitterEvent)
	return
}
