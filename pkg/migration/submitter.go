package migration

import (
	"errors"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/elchead/k8s-migration-controller/pkg/migration"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ControllerI interface {
	GetMigrations() (migrations []migration.MigrationCmd, err error)
}

type MigrationSubmitter struct {
	controller ControllerI
	jobs []jobparser.PodMemory
}

func (m *MigrationSubmitter) Submit(
	currentTime clock.Clock,
	n algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	migrations, err := m.controller.GetMigrations()
	if err != nil {
		return nil, err
	}

	events := make([]submitter.Event, 0, len(migrations)+1)
	for _,cmd := range migrations {
		job := jobparser.GetJob(cmd.Pod,m.jobs)
		if job == nil {
			return nil,errors.New("could not get job")
		}
		events = append(events, &submitter.SubmitEvent{Pod:jobparser.MigratePod(*job,currentTime.ToMetaV1().Time)})
	}
	// }
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
	return &MigrationSubmitter{controller: controller,jobs: jobs}
}

