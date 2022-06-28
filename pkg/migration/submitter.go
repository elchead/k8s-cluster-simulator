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
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
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

func NewConcurrentMigrationChecker() *concurrentMigrationChecker {
	return &concurrentMigrationChecker{make(map[string]clock.Clock),make(map[string]clock.Clock),nil}
}

type concurrentMigrationChecker struct {
	migrationFinish map[string]clock.Clock
	migrationStart map[string]clock.Clock
	latestFinish *clock.Clock
}
func (m *concurrentMigrationChecker) StartMigration(t clock.Clock,gbSize float64,pod string) {
	m.migrationStart[pod] = t
	m.migrationFinish[pod] = m.getLastMigrationFinishTime(gbSize,t)
}

func (m *concurrentMigrationChecker) getLastMigrationFinishTime(gbSize float64,now clock.Clock) clock.Clock {
	if m.latestFinish == nil {
		m.latestFinish = &now
	}

	startTime := maxClock(*m.latestFinish, now)
	res := startTime.Add(GetMigrationTime(gbSize))
	m.latestFinish = &res
	return res
}

func maxClock(t1, t2 clock.Clock) (startTime clock.Clock) {
	if t2.BeforeOrEqual(t1) {
		startTime = t1
	} else {
		startTime = t2
	}
	return
}

func (m *concurrentMigrationChecker) GetMigrationFinishTime(pod string) clock.Clock {
	return m.migrationFinish[pod]
} 

func (m *concurrentMigrationChecker) IsReady(current clock.Clock) bool { return true }

func NewMigrationChecker(checkerType string) MigrationCheckerI {
	switch checkerType {
	case "blocking":  return NewBlockingMigrationChecker()
	case "concurrent": return NewConcurrentMigrationChecker()
	default: 
		log.L.Warnf("unsupported checker type %v; using blocking type", checkerType) 
		return &blockingMigrationChecker{MigrationChecker{}}
	}
}
type blockingMigrationChecker struct {
	adapter MigrationChecker
}

func NewBlockingMigrationChecker() *blockingMigrationChecker {
	return &blockingMigrationChecker{MigrationChecker{}}
}

func (m *blockingMigrationChecker) StartMigration(t clock.Clock,gbSize float64,pod string) {
	m.adapter.StartMigrationWithSize(t,gbSize)	
}

func (m *blockingMigrationChecker) GetMigrationFinishTime(pod string) clock.Clock {
	return m.adapter.GetMigrationFinishTime()
} 

func (m *blockingMigrationChecker) IsReady(current clock.Clock) bool { return m.adapter.IsReady(current) }


type MigrationCheckerI interface {
	StartMigration(t clock.Clock,gbSize float64,pod string)
	GetMigrationFinishTime(pod string) clock.Clock
	IsReady(current clock.Clock) bool
}


type MigrationChecker struct {
	migrationStart clock.Clock
	migrationDuration time.Duration
}

func (m *MigrationChecker) StartMigration(t clock.Clock) {
	m.migrationStart = t
	m.migrationDuration = MigrationTime
}

func (m *MigrationChecker) StartMigrationWithSize(t clock.Clock,gbSize float64)  {
	m.migrationStart = t
	m.migrationDuration = GetMigrationTime(gbSize)
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
	checker monitoring.MigrationCheckerI
	factory jobparser.PodFactory
}

func (m *MigrationSubmitter) Submit(
	currentTime clock.Clock,
	n algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	var freezevents []submitter.Event
	if m.checker.IsReady(currentTime) {
		// log.L.Infof("migration checker is ready")
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
	migevents := m.getEventsFromFinishedMigrations(currentTime)
	events := append(freezevents, migevents...)
	// terminate
	isSimulationFinished := m.endTime.BeforeOrEqual(currentTime)
	if isSimulationFinished {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}
	return events, nil
}

func (m *MigrationSubmitter) getEventsFromFinishedMigrations(currentTime clock.Clock) []submitter.Event {
	events := make([]submitter.Event, 0, m.queue.RemainingValues()+1)
	for m.queue.ExistNext() || m.queue.RemainingValues() == 1 {
		nextJob := m.queue.Value()
		jobTime := clock.NewClock(nextJob.StartAt)
		if jobTime.BeforeOrEqual(currentTime) {
			log.L.Debug("pop from queue:", nextJob.Name)
			job := jobparser.GetJob(nextJob.Name, m.jobs) // jobs are copied in iterator.. need original reference for communication with job deleter
			jobparser.UpdateJobNameForMigration(job) // update global job reference name only when migration is finished
			jobparser.UpdateJobNameForMigration(&nextJob) // TODO improve design
			pod := m.factory.NewMigratedPodToNode(nextJob)
			nextJob.FinishedMigration()
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
		if cmd.NewNode == "" {
			log.L.Info("No migrating node set for ", jobName)
		}
		job.IsMigratingToNode = cmd.NewNode // true

		podsize := cmd.Usage // GB
		m.checker.StartMigration(currentTime,podsize,jobName) // TODO use info of new node ?
		finishTime :=  m.checker.GetMigrationFinishTime(jobName).ToMetaV1().Time
		jobparser.UpdateJobForMigration(job,currentTime.ToMetaV1().Time)
		log.L.Debug("push migration to queue:", job.Name, " size ",podsize, " to node ",job.IsMigratingToNode," finishing at ", finishTime)
		m.queue.Push(job)
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
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),factory: jobparser.PodFactory{SetResources:false},checker: NewBlockingMigrationChecker()}
}

func NewSubmitterWithJobsWithEndTimeFactory(controller ControllerI,jobs []jobparser.PodMemory,endTime time.Time,factory jobparser.PodFactory,checker monitoring.MigrationCheckerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),factory: factory,checker: checker}
}

