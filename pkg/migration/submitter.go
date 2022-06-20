package migration

import (
	"fmt"
	"math"
	"time"

	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/elchead/k8s-cluster-simulator/pkg/util"
	"github.com/elchead/k8s-migration-controller/pkg/migration"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

const MigrationTime = 5 * time.Minute
const BackoffInterval = 45 * time.Second
type ControllerI interface {
	GetMigrations() (migrations []migration.MigrationCmd, err error)
}

func GetMigrationTime(gbSz float64) time.Duration {
	return time.Duration(math.Ceil(3.3506*gbSz))*time.Second
}

type MigrationChecker struct {
	migrationStart clock.Clock
	migrationDuration time.Duration
}

func (m *MigrationChecker) StartMigration(t clock.Clock) {
	m.migrationStart = t
	m.migrationDuration = MigrationTime
}

func (m *MigrationChecker) StartMigrationWithSize(t clock.Clock,size float64)  {
	m.migrationStart = t
	m.migrationDuration = GetMigrationTime(size)
}

func (m *MigrationChecker) GetMigrationFinishTime() clock.Clock {
	return m.migrationStart.Add(m.migrationDuration)
}

func (m *MigrationChecker) IsReady(current clock.Clock) bool { return m.GetMigrationFinishTime().Add(BackoffInterval).BeforeOrEqual(current) } 

type MigrationSubmitter struct {
	controller ControllerI
	jobs []jobparser.PodMemory
	queue jobparser.Iterator
	endTime clock.Clock
	checker MigrationChecker
	factory jobparser.PodFactory
}

func (m *MigrationSubmitter) Submit(
	currentTime clock.Clock,
	n algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	var freezevents []submitter.Event
	if m.checker.IsReady(currentTime) {
		migrations, err := m.controller.GetMigrations()
		if err != nil {
			return []submitter.Event{}, errors.Wrap(err, "migrator failed")
		}
		err = m.startMigrations(migrations, currentTime)
		if err != nil {
			return []submitter.Event{}, errors.Wrap(err, "failed to start migration")
		}

		freezevents = make([]submitter.Event, 0, len(migrations)+1)
		for _, cmd := range migrations {
			freezevents = append(freezevents, &submitter.FreezeUsageEvent{PodKey:cmd.Pod})
		}

	}
	migevents := m.getEventsFromMigrations(currentTime)
	events := append(freezevents, migevents...)
	
	// terminate
	isSimulationFinished := m.endTime.BeforeOrEqual(currentTime)
	if isSimulationFinished {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}
	return events, nil
}

func (m *MigrationSubmitter) getEventsFromMigrations(currentTime clock.Clock) []submitter.Event {
	events := make([]submitter.Event, 0, m.queue.RemainingValues()+1)
	for m.queue.ExistNext() || m.queue.RemainingValues() == 1 {
		nextJob := m.queue.Value()
		jobTime := clock.NewClock(nextJob.StartAt)
		if jobTime.BeforeOrEqual(currentTime) {
			log.L.Debug("pop from queue:", nextJob.Name)
			pod := m.factory.NewMigratedPod(nextJob)
			nextJob.IsMigrating = false
			events = append(events, &submitter.SubmitEvent{Pod: pod})

			oldPod := util.GetOldPodName(pod.Name)
			events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: oldPod})

			m.queue.Next()
		} else {
			break
		}
	}
	return events
}

func (m *MigrationSubmitter) startMigrations(migrations []migration.MigrationCmd, currentTime clock.Clock) error {
	for _, cmd := range migrations {
		jobName := util.PodNameWithoutNamespace(cmd.Pod)
		job := jobparser.GetJob(jobName, m.jobs)
		if job == nil {
			return fmt.Errorf("could not get job %s", jobName)
		}
		job.Name = jobName
		job.IsMigrating = true

		m.checker.StartMigration(currentTime)
		jobparser.UpdateJobForMigration(job, m.checker.GetMigrationFinishTime().ToMetaV1().Time)
		log.L.Debug("push to queue:", job.Name)
		m.queue.Push(*job)
	}
	return nil
}

func NewSubmitter(controller ControllerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller}
}

func NewSubmitterWithJobs(controller ControllerI,jobs []jobparser.PodMemory) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{})}
}

func NewSubmitterWithJobsWithEndTime(controller ControllerI,jobs []jobparser.PodMemory,endTime time.Time) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime)}
}

func NewSubmitterWithJobsWithEndTimeFactory(controller ControllerI,jobs []jobparser.PodMemory,endTime time.Time,factory jobparser.PodFactory) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),factory: factory}
}

