package migration_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	cmigration "github.com/elchead/k8s-migration-controller/pkg/migration"
	"github.com/stretchr/testify/assert"

	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
)

func TestSubmitter(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)
	
	t.Run("no migration cmd returns no events", func(t *testing.T) {
		controllerStub := EmptyController{}
		sut := migration.NewSubmitter(controllerStub) 
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)
		assert.Empty(t,events)
	})
	t.Run("migration cmd with name and usage returns pod with that spec", func(t *testing.T) {
		controllerStub := ControllerStub{"worker",1e9}
		sut := migration.NewSubmitter(controllerStub) 
		events, err := sut.Submit(simTime, nil, nil)
		assert.NoError(t, err)
		assertSubmitEvent(t, events[0], "worker")
	})

}

func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	submit, ok := event.(*submitter.SubmitEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, submit.Pod.ObjectMeta.Name)
}

type EmptyController struct {}

func (e EmptyController) GetMigrations() (migrations []cmigration.MigrationCmd, err error) {
	return []cmigration.MigrationCmd{}, nil
}

type ControllerStub struct {
	name string
	usage float64
}

func (e ControllerStub) GetMigrations() (migrations []cmigration.MigrationCmd, err error) {
	return []cmigration.MigrationCmd{{Pod:e.name,Usage:e.usage}}, nil
}



