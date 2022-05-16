package migration_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	cmigration "github.com/elchead/k8s-migration-controller/pkg/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubmitter(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)

	t.Run("migrate job after 5 minutes", func(t *testing.T) {
		jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
		controllerStub := ControllerStub{}
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"j2",Usage:1e9}}, nil).Once()
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"j1",Usage:1e9}}, nil).Once()
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()


		sut := migration.NewSubmitterWithJobs(controllerStub,jobs) 

		t.Run("migrate no job yet", func(t *testing.T) {
			events, err := sut.Submit(simTime, nil, nil)
			assert.NoError(t, err)
			assert.Empty(t, events)
		})
		
		afterMigration := now.Add(migration.MigrationTime)
		t.Run("migrate j2", func(t *testing.T) {
			events, err := sut.Submit(clock.NewClock(afterMigration), nil, nil)	
			assert.NoError(t, err)
			assertSubmitEvent(t, events[0], "mj2")
		})

		t.Run("migrate j1", func(t *testing.T) {
			afterMigration = afterMigration.Add(migration.MigrationTime)
			events, err := sut.Submit(clock.NewClock(afterMigration), nil, nil)
			assert.NoError(t, err)
			assertSubmitEvent(t, events[0], "mj1")
		})
	})

}

func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	submit, ok := event.(*submitter.SubmitEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, submit.Pod.ObjectMeta.Name)
}
type ControllerStub struct {
	mock.Mock
}

func (c ControllerStub) GetMigrations() (migrations []cmigration.MigrationCmd, err error) {
	args := c.Called()
	return args.Get(0).([]cmigration.MigrationCmd), args.Error(1)
}



