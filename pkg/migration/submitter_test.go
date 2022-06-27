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
	controllerStub.On("GetMigrations").Return([]cmigration.MigrationCmd{{Pod:"default/j2",Usage:30}}, nil).Once() // 10GB
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
		afterMigration := clockNow.Add(migration.MigrationTime+migration.BackoffInterval)
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
		migrationTime := migration.GetMigrationTime(10)
		beforeBackOff := clockNow.Add(migrationTime + 20 * time.Second)
		sut.Submit(beforeBackOff,nil, nil)	
		controllerStub.AssertNumberOfCalls(suite.T(), "GetMigrations",2)
	})

	suite.Run("call controller after backoff", func(){
		afterBackOff := clockNow.Add(migration.MigrationTime + migration.BackoffInterval)
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
	afterMigration := submissionTime.Add(migration.MigrationTime)
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

func TestCheckerMigrationProcess(t *testing.T) {
	sut := migration.MigrationChecker{}
	t.Run("not ready during migration", func(t *testing.T){
		sut.StartMigration(clockNow)
		assert.False(t,sut.IsReady(clockNow.Add(migration.BackoffInterval)))
	})
	t.Run("not ready before backoff", func(t *testing.T){
		assert.False(t,sut.IsReady(clockNow.Add(1*time.Second)))
	})
	t.Run("ready after backoff", func(t *testing.T){
		assert.True(t,sut.IsReady(clockNow.Add(migration.MigrationTime + migration.BackoffInterval)))
	})
}

func TestCheckerConcurrentMigration(t *testing.T) {
	sut := migration.NewConcurrentMigrationChecker()
	now := clock.NewClock(time.Now())
	sut.StartMigration(now,10.,"pod1")
	assert.True(t,sut.IsReady(now.Add(1* time.Second)))
	end := sut.GetMigrationFinishTime("pod1")
	migrationTime := end.Sub(now)
	sut.StartMigration(now,20.,"pod2")
	assertTimeRoughlyEqual(t,now.Add(3*migrationTime),sut.GetMigrationFinishTime("pod2"))
	t.Run("much later migration starts much later",func(t *testing.T){
		later := now.Add(5*time.Hour)
		sut.StartMigration(later,10.,"pod3")
		assertTimeRoughlyEqual(t,later.Add(1*migrationTime),sut.GetMigrationFinishTime("pod3"))
	})
}

func assertTimeRoughlyEqual(t testing.TB,time1 clock.Clock, time2 clock.Clock) {
	assert.Equal(t,time1.ToMetaV1().Time.Round(1*time.Second),time2.ToMetaV1().Time.Round(1*time.Second))	
}

func TestGetMigrationTime(t *testing.T) {
	assert.Equal(t,168*time.Second,migration.GetMigrationTime(50))
}
