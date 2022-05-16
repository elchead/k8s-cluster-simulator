package migration

import (
	"errors"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/elchead/k8s-migration-controller/pkg/migration"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func (m *MigrationSubmitter) Submit(
	currentTime clock.Clock,
	n algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	migrations, err := m.controller.GetMigrations()
	if err != nil {
		return nil, err
	}

	// add migrations to queue
	for _,cmd := range migrations {
		job := jobparser.GetJob(cmd.Pod,m.jobs)
		if job == nil {
			return nil,errors.New("could not get job")
		}
		migrationTime := currentTime.ToMetaV1().Time.Add(MigrationTime)
		migratedJob := jobparser.UpdateJobForMigration(*job,migrationTime)

		m.queue.Push(migratedJob)
	}

	// check queue and add events
	events := make([]submitter.Event, 0, m.queue.RemainingValues()+1)
	for m.queue.ExistNext() || m.queue.RemainingValues() == 1 {
		nextJob := m.queue.Value()
		jobTime := clock.NewClock(nextJob.StartAt)
		if jobTime.BeforeOrEqual(currentTime) {
			pod := jobparser.CreatePod(nextJob)
			events = append(events, &submitter.SubmitEvent{Pod: pod})
			m.queue.Next()
		} else {
			break
		}
	}
	return events, err
}

func newPod(name string,memUsage float64) *v1.Pod {
	simSpec := ""
// 	for i := 0; i < s.myrand.Intn(4)+1; i++ {
// 		sec := 60 * s.myrand.Intn(60)
// 		cpu := 1 + s.myrand.Intn(4)
// 		mem := 1 + s.myrand.Intn(4)
// 		gpu := s.myrand.Intn(2)

// 		simSpec += fmt.Sprintf(`
// - seconds: %d
//   resourceUsage:
//     cpu: %d
//     memory: %dGi
//     nvidia.com/gpu: %d
// `, sec, cpu, mem, gpu)
// 	}

	// prio := s.myrand.Int31n(3) / 2 // 0, 0, 1

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Annotations: map[string]string{
				"simSpec": simSpec,
			},
		},
		// Spec: v1.PodSpec{
		// 	Containers: []v1.Container{
		// 		{
		// 			Name:  "container",
		// 			Image: "container",
		// 			Resources: v1.ResourceRequirements{
		// 				Requests: v1.ResourceList{
		// 					"cpu":            resource.MustParse("4"),
		// 					"memory":         resource.MustParse("4Gi"),
		// 					"nvidia.com/gpu": resource.MustParse("1"),
		// 				},
		// 				Limits: v1.ResourceList{
		// 					"cpu":            resource.MustParse("6"),
		// 					"memory":         resource.MustParse("6Gi"),
		// 					"nvidia.com/gpu": resource.MustParse("1"),
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}

	return &pod
}

func NewSubmitter(controller ControllerI) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller}
}

func NewSubmitterWithJobs(controller ControllerI,jobs []jobparser.PodMemory) *MigrationSubmitter {
	return &MigrationSubmitter{controller: controller,jobs: jobs,queue: *jobparser.NewIterator([]jobparser.PodMemory{})}
}

