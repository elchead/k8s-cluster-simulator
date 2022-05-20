package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetMemU(t *testing.T){
spec :=`
- seconds: 0.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 60.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 120.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 180.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 240.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 300.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 360.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 420.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 480.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 540.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 600.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 660.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 720.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 780.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 840.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 900.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 960.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1020.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1080.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1140.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1200.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1260.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1320.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1380.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1440.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1500.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1560.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1620.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1680.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1740.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1800.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1860.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1920.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 1980.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2040.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2100.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2160.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2220.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2280.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2340.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2400.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2460.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2520.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2580.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2640.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2700.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2760.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2820.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2880.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000

- seconds: 2940.000000
  resourceUsage:
    cpu: 8
    memory: 35430400.000000

- seconds: 3000.000000
  resourceUsage:
    cpu: 8
    memory: 38672384.000000

- seconds: 3060.000000
  resourceUsage:
    cpu: 8
    memory: 3998681088.000000

- seconds: 3120.000000
  resourceUsage:
    cpu: 8
    memory: 35107024896.000000

- seconds: 3180.000000
  resourceUsage:
    cpu: 8
    memory: 49950025728.000000

- seconds: 3240.000000
  resourceUsage:
    cpu: 8
    memory: 64947959125.000000

- seconds: 3300.000000
  resourceUsage:
    cpu: 8
    memory: 79159287808.000000

- seconds: 3360.000000
  resourceUsage:
    cpu: 8
    memory: 88796837205.000000

- seconds: 3420.000000
  resourceUsage:
    cpu: 8
    memory: 90342062080.000000

- seconds: 3480.000000
  resourceUsage:
    cpu: 8
    memory: 90341859328.000000

- seconds: 3540.000000
  resourceUsage:
    cpu: 8
    memory: 90341588992.000000

- seconds: 3600.000000
  resourceUsage:
    cpu: 8
    memory: 90857142272.000000

- seconds: 3660.000000
  resourceUsage:
    cpu: 8
    memory: 96237842432.000000

- seconds: 3720.000000
  resourceUsage:
    cpu: 8
    memory: 104524275712.000000

- seconds: 3780.000000
  resourceUsage:
    cpu: 8
    memory: 117007682901.000000

- seconds: 3840.000000
  resourceUsage:
    cpu: 8
    memory: 141040237909.000000

- seconds: 3900.000000
  resourceUsage:
    cpu: 8
    memory: 155942408192.000000

- seconds: 3960.000000
  resourceUsage:
    cpu: 8
    memory: 163195846656.000000

- seconds: 4020.000000
  resourceUsage:
    cpu: 8
    memory: 163564692138.000000

- seconds: 4080.000000
  resourceUsage:
    cpu: 8
    memory: 167729077589.000000

- seconds: 4140.000000
  resourceUsage:
    cpu: 8
    memory: 206899395925.000000

- seconds: 4200.000000
  resourceUsage:
    cpu: 8
    memory: 263484478122.000000

- seconds: 4260.000000
  resourceUsage:
    cpu: 8
    memory: 281264732160.000000

- seconds: 4320.000000
  resourceUsage:
    cpu: 8
    memory: 281242525696.000000

- seconds: 4380.000000
  resourceUsage:
    cpu: 8
    memory: 298349284010.000000

- seconds: 4440.000000
  resourceUsage:
    cpu: 8
    memory: 298040209408.000000

- seconds: 4500.000000
  resourceUsage:
    cpu: 8
    memory: 252056158208.000000

- seconds: 4560.000000
  resourceUsage:
    cpu: 8
    memory: 238594512896.000000

- seconds: 4620.000000
  resourceUsage:
    cpu: 8
    memory: 246699758933.000000

- seconds: 4680.000000
  resourceUsage:
    cpu: 8
    memory: 179591500458.000000

- seconds: 4740.000000
  resourceUsage:
    cpu: 8
    memory: 163978614784.000000

- seconds: 4800.000000
  resourceUsage:
    cpu: 8
    memory: 163978704896.000000

- seconds: 4860.000000
  resourceUsage:
    cpu: 8
    memory: 163982813866.000000

- seconds: 4920.000000
  resourceUsage:
    cpu: 8
    memory: 160643034453.000000

- seconds: 4980.000000
  resourceUsage:
    cpu: 8
    memory: 159972055040.000000

- seconds: 5040.000000
  resourceUsage:
    cpu: 8
    memory: 159971942400.000000

- seconds: 5100.000000
  resourceUsage:
    cpu: 8
    memory: 159973384192.000000

- seconds: 5160.000000
  resourceUsage:
    cpu: 8
    memory: 133239389525.000000

- seconds: 5220.000000
  resourceUsage:
    cpu: 8
    memory: 0.000000
`
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "default",
			Annotations: map[string]string{"simSpec": spec },
		}}
	specO, err := parseSpec(pod)
	assert.NoError(t,err)
	assert.Equal(t,35430400000000.,GetPodUsage(specO,2900)["memory"])

}
