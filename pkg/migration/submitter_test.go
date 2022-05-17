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
	endTime := now.Add(30 * time.Minute)

	t.Run("migrate job after 5 minutes", func(t *testing.T) {
		jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
		controllerStub := new(ControllerStub)
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil).Once()
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j1",Usage:1e9}}, nil).Once()
		controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()


		sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 

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
			after := afterMigration.Add(2* time.Second) 
			sut.Submit(clock.NewClock(after), nil, nil)	// call submit to issue cmd (was blocked before)
			
			afterMigration = after.Add(migration.MigrationTime)
			events, err := sut.Submit(clock.NewClock(afterMigration), nil, nil)
			assert.NoError(t, err)
			assertSubmitEvent(t, events[0], "mj1")
		})
	})
}

func TestMigrateMigratedJob(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)
	endTime := now.Add(30 * time.Minute)

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/mj2",Usage:1e9}}, nil)	
	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 
	_, err := sut.Submit(simTime, nil, nil)
	assert.NoError(t, err)


	afterMigration := clock.NewClock(now.Add(migration.MigrationTime))
	events, err := sut.Submit(afterMigration, nil, nil)	
	assert.NoError(t, err)
	assertSubmitEvent(t, events[0], "mmj2") 
}

func TestNoRepeatedCallsUntilMigrationFinished(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)
	endTime := now.Add(30 * time.Minute)

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil)
	
	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 
	sut.Submit(simTime, nil, nil)
	sut.Submit(simTime, nil, nil)
	controllerStub.AssertNumberOfCalls(t, "GetMigrations", 1)

}

func assertTerminateEvent(t testing.TB, event submitter.Event) {
	assert.True(t, isTerminateEvent(event))
}

func isTerminateEvent(event submitter.Event) (ok bool) {
	_, ok = event.(*submitter.TerminateSubmitterEvent)
	return
}


func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	submit, ok := event.(*submitter.SubmitEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, submit.Pod.ObjectMeta.Name)
}
type ControllerStub struct {
	mock.Mock
}

func (c *ControllerStub) GetMigrations() (migrations []cmigration.MigrationCmd, err error) {
	args := c.Called()
	return args.Get(0).([]cmigration.MigrationCmd), args.Error(1)
}



