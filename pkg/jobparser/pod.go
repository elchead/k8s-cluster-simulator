package jobparser

import (
	"fmt"
	"strings"
	"time"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Time interface {
	Before(u Time) bool
	After(u Time) bool
}

type PodFactory struct {
	SetResources bool
}

func (f PodFactory)  New(podinfo PodMemory) *v1.Pod {
	if f.SetResources {
		return CreatePod(podinfo)
	} else {
		return CreatePodWithoutResources(podinfo)
	}
}

func (f PodFactory)  NewMigratedPod(podinfo PodMemory) *v1.Pod {
	// TODO set parameter for increase
	return f.NewWithResources(podinfo,fmt.Sprintf(`%f`,1. *podinfo.Records[0].Usage))
}

func (f PodFactory)  NewWithResources(podinfo PodMemory,memSize string) *v1.Pod {
	pod := CreatePodWithoutResources(podinfo)
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "worker",
				Image: "worker-image",
				Resources: GetPodRequest(memSize),
			},
		},
	}
	return pod
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

func UpdateJobForMigration(podinfo *PodMemory, migration time.Time) {

	podinfo.Records = FilterRecordsBefore(podinfo.Records,migration)
	// podinfo.Name = "m" + podinfo.Name
	podinfo.StartAt = migration	
}

func UpdateJobNameForMigration(podinfo *PodMemory) {
	podinfo.Name = "m" + podinfo.Name
}

func GetJobSizeFromName(name string) (string, error) {
	s := strings.Split(name, "-")
	if s[0] == name || len(s) < 2 {
		return "", fmt.Errorf("job size of %s could not be deduced",name)
	} else {
		return s[2],nil
	}
}

func CreatePodWithoutResources(podinfo PodMemory) *v1.Pod {
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
	return &v1.Pod{
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
	}
}

func CreatePod(podinfo PodMemory) *v1.Pod {
	size,err := GetJobSizeFromName(podinfo.Name)
	if err != nil {
		log.L.Info("Setting job size to s since:",err)
		size = "s"
	}
	pod := CreatePodWithoutResources(podinfo)
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "worker",
				Image: "worker-image",
				Resources: GetJobResources(size),
			},
		},
	}
	return pod
}

func GetJobResourceRequest(size string) v1.ResourceList {
	switch size {
	case "s":
		return v1.ResourceList{
			"cpu":            resource.MustParse("5"),
			"memory":         resource.MustParse("30Gi"),
		      }
	case "m":
		return 	v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("80Gi"),
		      }
	case "l": return v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("130Gi"),
		      }

	case "xl": return v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("420Gi"),
		      }
	default:
		return v1.ResourceList{}
	}
}

func GetJobResourceLimit() v1.ResourceList {
	return v1.ResourceList{
		"cpu":            resource.MustParse("10"),
		"memory":         resource.MustParse("430Gi"),
	      }
}

func GetJobResources(size string) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: GetJobResourceRequest(size),
		Limits: GetJobResourceLimit(),
	}
}

func GetPodRequest(memSize string) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			"cpu":            resource.MustParse("5"),
			"memory":         resource.MustParse(memSize),
		      },
		Limits: GetJobResourceLimit(),
	}
}
