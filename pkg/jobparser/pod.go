package jobparser

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FilterRecordsBefore(podmem []Record, t time.Time) []Record {
	res := make([]Record, 0, len(podmem))
	var beforeIdx int
	for i, record := range podmem {
		if !record.Time.Before(t) {
			beforeIdx = i
		}
	}
	
	// check between
	if podmem[beforeIdx-1].Time.Before(t) && podmem[beforeIdx].Time.After(t) {
		podmem[beforeIdx-1].Time = t
		res = append(res,podmem[beforeIdx-1])
	}

	return append(res, podmem[beforeIdx:]...)
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
