package migration

import (
	"time"

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
type ControllerI interface {
	GetMigrations() (migrations []migration.MigrationCmd, err error)
}

type MigrationSubmitter struct {
	controller ControllerI
	jobs []jobparser.PodMemory
	queue jobparser.Iterator
	endTime clock.Clock
	migrationInProcess bool
}

func (m *MigrationSubmitter) Submit(
	currentTime clock.Clock,
	n algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	if !m.migrationInProcess {
		migrations, err := m.controller.GetMigrations()
		if err != nil {
			return []submitter.Event{}, errors.Wrap(err, "failed to get migrations")
		}
		
		// add migrations to queue
		for _,cmd := range migrations {
			jobName :=  util.PodNameWithoutNamespace(cmd.Pod) //util.JobNameFromPod(cmd.Pod)
			job := jobparser.GetJob(jobName,m.jobs)
			if job == nil {
				return nil,errors.New("could not get job "+ jobName)
			}
			job.Name = jobName

			migrationTime := currentTime.ToMetaV1().Time.Add(MigrationTime)
			jobparser.UpdateJobForMigration(job,migrationTime)
	
			m.queue.Push(*job)
			m.migrationInProcess = true
		}
	}

	// check queue and add events
	events := make([]submitter.Event, 0, m.queue.RemainingValues()+1)
	for m.queue.ExistNext() || m.queue.RemainingValues() == 1 {
		nextJob := m.queue.Value()
		jobTime := clock.NewClock(nextJob.StartAt)
		if jobTime.BeforeOrEqual(currentTime) {
			pod := jobparser.CreatePod(nextJob)
			events = append(events, &submitter.SubmitEvent{Pod: pod})

			oldPod := util.GetOldPodName(pod.Name)
			events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: oldPod})
			
			m.queue.Next()
			m.migrationInProcess = false
		} else {
			break
		}
	}

	// terminate
	if m.endTime.BeforeOrEqual(currentTime) {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}
	return events, nil
}

func NewSubmitter(controller ControllerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller}
}

func NewSubmitterWithJobs(controller ControllerI,jobs []jobparser.PodMemory) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{})}
}

func NewSubmitterWithJobsWithEndTime(controller ControllerI,jobs []jobparser.PodMemory,endTime time.Time) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{}),endTime: clock.NewClock(endTime),migrationInProcess: false}
}

