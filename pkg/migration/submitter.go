package migration

import (
	"fmt"
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

type MigrationSubmitter struct {
	controller monitoring.ControllerI
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
		migrations,migerr := m.controller.GetMigrations()
		var nodeErr *monitoring.NodeFullError
		if migerr!=nil && !errors.As(migerr,&nodeErr) {
			return []submitter.Event{}, errors.Wrap(migerr, "migrator failed")
		} else if errors.As(migerr,&nodeErr){
			log.L.Info("Provision new node: ", migerr.Error())
		}
		err := m.startMigrations(migrations, currentTime)
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
		jobparser.UpdateJobForMigration(job,currentTime.ToMetaV1().Time,finishTime)
		log.L.Debug("push migration to queue:", job.Name, " size ",podsize, " to node ",job.IsMigratingToNode," finishing at ", finishTime)
		log.L.Infof("MigrationTime %s %.0f starting %s", jobName, finishTime.Sub(currentTime.ToMetaV1().Time).Seconds(),currentTime)
		m.queue.Push(job)
	}
	return nil
}

func NewSubmitter(controller monitoring.ControllerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller}
}

func NewSubmitterWithJobs(controller  monitoring.ControllerI,jobs []jobparser.PodMemory) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{})}
}

func NewSubmitterWithJobsWithEndTime(controller monitoring.ControllerI,jobs []jobparser.PodMemory,endTime time.Time) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),factory: jobparser.PodFactory{SetResources:false},checker: monitoring.NewBlockingMigrationChecker()}
}

func NewSubmitterWithJobsWithEndTimeFactory(controller  monitoring.ControllerI,jobs []jobparser.PodMemory,endTime time.Time,factory jobparser.PodFactory,checker monitoring.MigrationCheckerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),factory: factory,checker: checker}
}

