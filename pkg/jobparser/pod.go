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
	RequestFactor float64
}

func NewPodFactory(requestFactor float64) PodFactory {
	setResources := requestFactor != 0.
	return PodFactory{setResources,requestFactor}
}

func (f PodFactory)  New(podinfo PodMemory) *v1.Pod {
	if f.SetResources {
		return CreatePodWithLimitsAndReq(podinfo,f.RequestFactor)
	} else {
		return CreatePodWithLimitsAndReq(podinfo,0.)
	}
}

func (f PodFactory)  NewMigratedPod(podinfo PodMemory) *v1.Pod {
	// TODO set parameter for increase
	return f.NewWithResources(podinfo,fmt.Sprintf(`%f`,1. *podinfo.Records[0].Usage))
}

func (f PodFactory)  NewMigratedPodToNode(podinfo PodMemory) *v1.Pod {
	res := f.NewMigratedPod(podinfo)
	res.Spec.NodeName = podinfo.IsMigratingToNode
	return res

}

func (f PodFactory)  NewWithResources(podinfo PodMemory,memSize string) *v1.Pod {
	pod := CreatePodWithLimitsAndReq(podinfo,0.) //f.New(podinfo)//CreatePodWithoutResources(podinfo)
	pod.Spec.Containers[0].Resources.Requests["memory"] = resource.MustParse(memSize)
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

// is called at starting time of migration. name is updated when migration finished
func UpdateJobForMigration(podinfo *PodMemory, migrationStart,migrationFinish time.Time) {

	podinfo.Records = FilterRecordsBefore(podinfo.Records,migrationStart)
	podinfo.StartAt = migrationFinish	
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

func CreatePodWithLimitsAndReq(podinfo PodMemory,memRequestFactor float64) *v1.Pod {
	size,err := GetJobSizeFromName(podinfo.Name)
	if err != nil {
		log.L.Info("Setting job size to s since: ",err)
		size = "s"
	}
	pod := CreatePodWithoutResources(podinfo)
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "worker",
				Image: "worker-image",
				Resources: GetJobResourcesWithRequest(GetJobResourceRequestWithFactor(size,memRequestFactor)),
			},
		},
	}
	return pod
}

func GetJobResourceRequest(size string) v1.ResourceList {
	return GetJobResourceRequestWithFactor(size,1.)	
}

func GetJobResourceRequestWithFactor(size string,factor float64) v1.ResourceList {
	switch size {
	case "s":
		return v1.ResourceList{
			"cpu":            resource.MustParse("5"),
			"memory":         resource.MustParse(getFractionalGi(30.,factor)),
		      }
	case "m":
		return 	v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse(getFractionalGi(80.,factor)),
		      }
	case "l": return v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse(getFractionalGi(130.,factor)),
		      }

	case "xl": return v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("420Gi"),
		      }
	default:
		return v1.ResourceList{}
	}
}

func getFractionalGi(amount,factor float64) string {
	return fmt.Sprintf("%.0fGi", factor*amount)
}

func GetJobResourceLimit() v1.ResourceList {
	// cannot set limit in practice because Kubernetes implicitly sets request to limit if no request is specified https://kubernetes.io/docs/tasks/administer-cluster/manage-resources/memory-default-namespace/#what-if-you-specify-a-container-s-limit-but-not-its-request // however simulator does not enforce thiss
	return v1.ResourceList{
		"cpu":            resource.MustParse("10"),
		"memory":         resource.MustParse("430Gi"),
	      }
}

func GetJobResources(size string) v1.ResourceRequirements {
	return GetJobResourcesWithRequest(GetJobResourceRequestWithFactor(size,1.))
}

func GetJobResourcesWithRequest(req v1.ResourceList) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests:req,
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
