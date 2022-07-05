package migration_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	cmigration "github.com/elchead/k8s-migration-controller/pkg/migration"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)
const MigrationTime = 5 * time.Minute
const BackoffInterval = 45 * time.Second

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


func (suite *MigrationSuite) TestSetNewNode() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:30,NewNode: "j3"}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()
	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
 	sut.Submit(clockNow, nil, nil)	
	events := assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
	assertContainsSubmitMigrationPodEventOnTargetNode(suite.T(),events,"mj2","j3")
	
}

func (suite *MigrationSuite) TestMigrateMultipleJobs() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:30}}, nil).Once() // GB
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j1",Usage:30}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
	suite.Run("do not issue migration pod before migration time finished", func() {
		events, err := sut.Submit(clockNow, nil, nil)
		assert.NoError(suite.T(), err)
		assertNoPodEvent(suite.T(), events)
	})
	suite.Run("do not call migration controller while migration in progress", func(){
		sut.Submit(clockNow, nil, nil)
		sut.Submit(clockNow.Add(1*time.Second), nil, nil)
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 1)	
	})
	
	suite.Run("migration pod is issued after migration time", func() {
		assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 2)
	})
	
	suite.Run("call migration controller again after migration + backoff and migrate new job", func() {
		afterMigration := clockNow.Add(MigrationTime+BackoffInterval)
		sut.Submit(afterMigration, nil, nil)
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations", 2)
		assertJobMigratedAfterTime(suite.T(),afterMigration,sut,"mj1")
	})
}
func (suite *MigrationSuite) TestMigrateMigratedJob() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/mj2",Usage:20}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()

	mjobs := suite.jobs
	mjobs[1].Name = "mj2"
	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,mjobs,endTime) 
	_, err := sut.Submit(clockNow, nil, nil)
	assert.NoError(suite.T(), err)

	assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mmj2")
}

func (suite *MigrationSuite) TestBackoffIntervalAfterMigration() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:10}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
	sut.Submit(clockNow,nil, nil)
	controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations",1)

	assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
	controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations",2)

	suite.Run("do not call controller before backoff interval", func(){
		migrationTime := monitoring.GetMigrationTime(10)
		beforeBackOff := clockNow.Add(migrationTime + 20 * time.Second)
		sut.Submit(beforeBackOff,nil, nil)	
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations",2)
	})

	suite.Run("call controller after backoff", func(){
		afterBackOff := clockNow.Add(MigrationTime + BackoffInterval)
		sut.Submit(afterBackOff,nil, nil)	
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations",3)	
	})

}

func (suite *MigrationSuite) TestTerminateSubmitterAtEndTime() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime) 
	events, err := sut.Submit(clock.NewClock(endTime), nil, nil)	
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), events, &submitter.TerminateSubmitterEvent{})
}

func (suite *MigrationSuite) TestAfterMigration() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:10}}, nil).Once()
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{}, nil).Once()	

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime)
	ev,_ := sut.Submit(clockNow, nil, nil)
	assertNoPodEvent(suite.T(), ev)
	suite.Run("delete old pod", func(){
		assertJobMigratedAfterTime(suite.T(),clockNow,sut,"mj2")
	})
	suite.Run("update job name in shared slice to migration pod", func(){
		assert.Equal(suite.T(),"mj2",suite.jobs[1].Name)
	})
}

func (suite *MigrationSuite) TestFreezeUsage() {
	controllerStub := new(ControllerStub)
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:1e9}}, nil)

	sut := migration.NewSubmitterWithJobsWithEndTime(controllerStub,suite.jobs,endTime)
	suite.Run("issue freeze event for migration",func(){
		events, err := sut.Submit(clockNow, nil, nil)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(),events,&submitter.FreezeUsageEvent{PodKey: "default/j2"})
	})
}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}


func assertNoPodEvent(t *testing.T, events []submitter.Event) {
	for _,event := range events {
		assert.IsType(t, &submitter.FreezeUsageEvent{}, event)
	}
}

func assertJobMigratedAfterTime(t testing.TB, submissionTime clock.Clock, sut *migration.MigrationSubmitter, migratedPodName string) []submitter.Event {
	afterMigration := submissionTime.Add(MigrationTime)
	events, err := sut.Submit(afterMigration, nil, nil)
	assert.NoError(t, err)
	assertContainsSubmitMigrationPodEvent(t, events, migratedPodName)
	assertContainsDeleteOldPodEvent(t, events, migratedPodName)
	return events
}

func assertDeleteEvent(t testing.TB, event submitter.Event, podName string) {
	assert.IsType(t,&submitter.DeleteEvent{},event)
}


func assertSubmitEvent(t testing.TB, event submitter.Event, podName string) {
	assert.IsType(t,&submitter.SubmitEvent{},event)
	assert.Equal(t, podName, event.(*submitter.SubmitEvent).Pod.ObjectMeta.Name)
}

func assertContainsSubmitMigrationPodEvent(t testing.TB, events []submitter.Event, podName string) {
	isContained := false
	for _,event := range events {
		if pod, ok := event.(*submitter.SubmitEvent); ok {
			if pod.Pod.ObjectMeta.Name == podName {
				isContained = true
			}
		}
	}
	assert.True(t, isContained,"contains submit event for "+podName)
}

func assertContainsSubmitMigrationPodEventOnTargetNode(t testing.TB, events []submitter.Event, podName,targetNode string) {
	isContained := false

	for _,event := range events {
		if pod, ok := event.(*submitter.SubmitEvent); ok {
			if pod.Pod.ObjectMeta.Name == podName { 
				if pod.Pod.Spec.NodeName == targetNode {
					isContained = true
				} else {
					assert.Fail(t, "pod "+podName+" is not on target node but on "+pod.Pod.Spec.NodeName)
				}
			}
		}
	}
	assert.True(t, isContained,"contains submit event on node "+ targetNode +" for "+podName)
}

func assertContainsDeleteOldPodEvent(t testing.TB, events []submitter.Event, podName string) {
	assert.Contains(t,events,&submitter.DeleteEvent{PodName:podName[1:],PodNamespace:"default"})
}
type ControllerStub struct {
	mock.Mock
}

func (c *ControllerStub) GetMigrations() (migrations []cmigration.MigrationCmd, err error) {
	args := c.Called()
	return args.Get(0).([]cmigration.MigrationCmd), args.Error(1)
}

