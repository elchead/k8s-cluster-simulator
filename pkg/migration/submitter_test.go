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

var now = time.Now()
var clockNow = clock.NewClock(now)
var endTime = now.Add(30 * time.Minute)
var jobs  = []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}

func TestMigrateMultipleJobs(t *testing.T) {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j1",Usage:1e9}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 
	t.Run("do not issue migration pod before migration time finished", func(t *testing.T) {
		events, err := sut.Submit(clockNow, nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, events)
	})
	t.Run("do not call migration controller while migration in progress", func(t *testing.T){
		sut.Submit(clockNow, nil, nil)
		sut.Submit(clockNow.Add(2*time.Second), nil, nil)
		controllerStub.AssertNumberOfCalls(t, "GetMigrations", 1)	
	})
	
	t.Run("migration pod is issued after migration time", func(t *testing.T) {
		assertJobMigratedAfterTime(t,clockNow,sut,"mj2")
		controllerStub.AssertNumberOfCalls(t, "GetMigrations", 1)
	})
	
	t.Run("call migration controller again after migration finished and migrate new job", func(t *testing.T) {
		afterMigration := clockNow.Add(migration.MigrationTime+2*time.Second)
		sut.Submit(afterMigration, nil, nil)
		controllerStub.AssertNumberOfCalls(t, "GetMigrations", 2)
		assertJobMigratedAfterTime(t,afterMigration,sut,"mj1")
	})
}

func TestMigrateMigratedJob(t *testing.T) {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/mj2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 
	_, err := sut.Submit(clockNow, nil, nil)
	assert.NoError(t, err)

	assertJobMigratedAfterTime(t,clockNow,sut,"mmj2")
}

func TestTerminateSubmitterAtEndTime(t *testing.T) {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/mj2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,jobs,endTime) 
	events, err := sut.Submit(clock.NewClock(endTime), nil, nil)	
	assert.NoError(t, err)
	assertTerminateEvent(t,events[0])
}


func assertJobMigratedAfterTime(t testing.TB, submissionTime clock.Clock, sut *migration.MigrationSubmitter, podName string) {
	afterMigration := submissionTime.Add(migration.MigrationTime)
	events, err := sut.Submit(afterMigration, nil, nil)
	assert.NoError(t, err)
	assertSubmitEvent(t, events[0], podName)
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



