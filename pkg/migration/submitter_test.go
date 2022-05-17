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
	"github.com/stretchr/testify/suite"
)

var now = time.Now()
var clockNow = clock.NewClock(now)
var endTime = now.Add(30 * time.Minute)
var jobs  = []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}

type MigrationSuite struct {
	suite.Suite
	jobs []jobparser.PodMemory
}

func (suite *MigrationSuite) SetupTest() {
	suite.jobs = make([]jobparser.PodMemory, len(jobs))
	copy(suite.jobs, jobs)
}

func (suite *MigrationSuite) TestMigrateMultipleJobs() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j1",Usage:1e9}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
	suite.Run("do not issue migration pod before migration time finished", func() {
		events, err := sut.Submit(clockNow, nil, nil)
		assert.NoError(suite.T(), err)
		assert.Empty(suite.T(), events)
	})
	suite.Run("do not call migration controller while migration in progress", func(){
		sut.Submit(clockNow, nil, nil)
		sut.Submit(clockNow.Add(2*time.Second), nil, nil)
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 1)	
	})
	
	suite.Run("migration pod is issued after migration time", func() {
		assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 1)
	})
	
	suite.Run("call migration controller again after migration finished and migrate new job", func() {
		afterMigration := clockNow.Add(migration.MigrationTime+2*time.Second)
		sut.Submit(afterMigration, nil, nil)
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 2)
		assertJobMigratedAfterTime(suite.T(),afterMigration,sut,"mj1")
	})
}
func (suite *MigrationSuite) TestTestMigrateMigratedJob() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/mj2",Usage:1e9}}, nil)

	mjobs := suite.jobs
	mjobs[1].Name = "mj2"
	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,mjobs,endTime) 
	_, err := sut.Submit(clockNow, nil, nil)
	assert.NoError(suite.T(), err)

	assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mmj2")
}

func (suite *MigrationSuite) TestTestTerminateSubmitterAtEndTime() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
	events, err := sut.Submit(clock.NewClock(endTime), nil, nil)	
	assert.NoError(suite.T(), err)
	assertTerminateEvent(suite.T(),events[0])
}

func (suite *MigrationSuite) TestAfterMigration() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime)
	sut.Submit(clockNow, nil, nil)
	suite.Run("update job name in shared slice to migration pod", func(){
		assert.Equal(suite.T(),"mj2",suite.jobs[1].Name)
	})
	suite.Run("delete old pod", func(){
		events := assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
		assertDeleteEvent(suite.T(),events[1],"j2")
	})

}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}






func assertJobMigratedAfterTime(t testing.TB, submissionTime clock.Clock, sut *migration.MigrationSubmitter, migratedPodName string) []submitter.Event {
	afterMigration := submissionTime.Add(migration.MigrationTime)
	events, err := sut.Submit(afterMigration, nil, nil)
	assert.NoError(t, err)
	assertSubmitEvent(t, events[0], migratedPodName)
	return events
}



func assertTerminateEvent(t testing.TB, event submitter.Event) {
	assert.True(t, isTerminateEvent(event))
}

func isTerminateEvent(event submitter.Event) (ok bool) {
	_, ok = event.(*submitter.TerminateSubmitterEvent)
	return
}

func assertDeleteEvent(t testing.TB, event submitter.Event, podName string) {
	delete, ok := event.(*submitter.DeleteEvent)
	assert.True(t, ok)
	assert.Equal(t, podName, delete.PodName)
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



