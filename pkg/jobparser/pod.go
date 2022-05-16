package jobparser

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Time interface {
	Before(u Time) bool
	After(u Time) bool
}

func FilterRecordsBefore(podmem []Record, t time.Time) []Record {
	res := make([]Record,0)
	var beforeIdx int
	for i, record := range podmem {
		if !record.Time.Before(t) {
			beforeIdx = i
			break
		}
	}
	// check between
	if len(podmem) == 1 {
		podmem[0].Time = t
		return append(res,podmem[0])

	} else if beforeIdx > 0 {
		if podmem[beforeIdx-1].Time.Before(t) && podmem[beforeIdx].Time.After(t) {
			podmem[beforeIdx-1].Time = t
			res = append(res,podmem[beforeIdx-1])
		}
	}

	return append(res, podmem[beforeIdx:]...)
}

func UpdateJobForMigration(podinfo PodMemory, migration time.Time) PodMemory {

	podinfo.Records = FilterRecordsBefore(podinfo.Records,migration)
	podinfo.Name = "m" + podinfo.Name
	podinfo.StartAt = migration	
	return podinfo
}


func MigratePod(podinfo PodMemory,migration time.Time) *v1.Pod {
	migratedJob := UpdateJobForMigration(podinfo,migration)
	return CreatePod(migratedJob)
}

func CreatePod(podinfo PodMemory) *v1.Pod {
	simSpec := ""
	cpu := "8" // s: 5-10; m: 8-10; l:8-10
	startTime := podinfo.Records[0].Time
	for _, record := range podinfo.Records {
		time := record.Time.Sub(startTime).Seconds()
		simSpec += fmt.Sprintf(`
- seconds: %f
  resourceUsage:
    cpu: %s
    memory: %f
`, time, cpu, record.Usage)
	}
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podinfo.Name,
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
		// 					"cpu":    resource.MustParse(cpu),
		// 					"memory": resource.MustParse("4Gi"),
		// 				},
		// 				// Limits: v1.ResourceList{
		// 				// 	"cpu":            resource.MustParse("6"),
		// 				// 	"memory":         resource.MustParse("6Gi"),
		// 				// 	"nvidia.com/gpu": resource.MustParse("1"),
		// 				// },
		// 			},
		// 		},
		// 	},
		// },
	}

	return &pod
}
