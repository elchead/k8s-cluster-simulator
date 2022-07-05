package pod_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/pod"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


func TestGetMemU(t *testing.T){
spec :=`
- seconds: 0
  resourceUsage:
    cpu: 8
    memory: 0

- seconds: 60
  resourceUsage:
    cpu: 8
    memory: 0
- seconds: 5160
  resourceUsage:
    cpu: 8
    memory: 0
`
	podv1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "default",
			Annotations: map[string]string{"simSpec": spec },
		}}
	now := clock.NewClock(time.Now())
	podo,_ := pod.NewPod(podv1,now,pod.Ok,"node")
	res := podo.Metrics(now.Add(1*time.Hour))
	assert.Equal(t,int32(3600),res.ExecutedSeconds)
	assert.Equal(t,int32(5220),res.Runtime)
	// specO, err := parseSpec(podv1)
	// assert.NoError(t,err)
	// assert.Equal(t,resource.MustParse("35430400000000e-6"),GetPodUsage(specO,2900)["memory"])
  }


