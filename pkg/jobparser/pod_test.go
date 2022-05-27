package jobparser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestPodFactory(t *testing.T) {
	sut := PodFactory{SetResources: false}

	podmem := PodMemory{Name: "o10n-worker-m-zx8wp-n5", Records: []Record{{Time: time.Now(), Usage: 1e9}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e2}}}
	podspec := sut.New(podmem)
	assert.Empty(t,podspec.Spec.Containers)
}

func TestPodFactorySetMigratedResources(t *testing.T) {
	podmem := PodMemory{Name: "o10n-worker-m-zx8wp-n5", Records: []Record{{Time: time.Now(), Usage: 2e2}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e5}}}
	
	sut := PodFactory{SetResources: false}
	podspec := sut.NewMigratedPod(podmem)
	assert.NotEmpty(t,podspec.Spec.Containers[0].Resources.Requests["memory"])
}

func TestPodFactoryWithResources(t *testing.T) {
	sut := PodFactory{SetResources: false}

	podmem := PodMemory{Name: "o10n-worker-m-zx8wp-n5", Records: []Record{{Time: time.Now(), Usage: 1e9}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e2}}}
	podspec := sut.NewWithResources(podmem,"10Gi")
	assert.Empty(t,podspec.Spec.Containers)
}

func TestPodSpecFromPodMemory(t *testing.T) {
	now := time.Now()
	podmem := PodMemory{Name: "w1", Records: []Record{{Time: now, Usage: 1e9}, {Time: now.Add(2 * time.Minute), Usage: 1e2}}}
	podspec := CreatePodWithoutResources(podmem)
	assert.Equal(t, "\n- seconds: 0.000000\n  resourceUsage:\n    cpu: 8\n    memory: 1000000000.000000\n\n- seconds: 120.000000\n  resourceUsage:\n    cpu: 8\n    memory: 100.000000\n", podspec.Annotations["simSpec"])
}

func TestGetJobSizeFromName(t *testing.T) {
	sz, err := GetJobSizeFromName("o10n-worker-m-zx8wp-n5")
	assert.NoError(t, err)
	assert.Equal(t,"m",sz)
}

func TestAssignResourcesFromPodName(t *testing.T) {
	podmem := PodMemory{Name: "o10n-worker-m-zx8wp-n5", Records: []Record{{Time: time.Now(), Usage: 1e9}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e2}}}
	podspec := CreatePod(podmem)
	assert.Equal(t,GetJobResources("m"),podspec.Spec.Containers[0].Resources)
}

func TestMigrationPod(t *testing.T) {
	now := time.Now()
	podmem := PodMemory{Name: "w1", Records: []Record{{Time: now, Usage: 1e9}, {Time: now.Add(2 * time.Minute), Usage: 1},{Time: now.Add(4 * time.Minute), Usage: 1e2}}}
	migrationTime := now.Add(3 * time.Minute)
	podspec := MigratePod(podmem,migrationTime)
	assert.Equal(t, "mw1", podspec.Name)
	assert.Equal(t, "\n- seconds: 0.000000\n  resourceUsage:\n    cpu: 8\n    memory: 1.000000\n\n- seconds: 60.000000\n  resourceUsage:\n    cpu: 8\n    memory: 100.000000\n", podspec.Annotations["simSpec"])
}

func TestFilterRecords(t *testing.T) {
	now := time.Now()
	records := []Record{{Time:now,  Usage: 1e9}, {Time: now.Add(2 * time.Minute), Usage: 1e2},{Time: now.Add(4 * time.Minute), Usage: 1e3},{Time: now.Add(6 * time.Minute), Usage: 1e4}}
	t.Run("get records bigger or equal that time",func(t *testing.T){
		assert.Equal(t,records[2:], FilterRecordsBefore(records,now.Add(4 * time.Minute)))
	})
	t.Run("start from last record when time in between two timestamps and set first time to checktime",func(t *testing.T){
		checkTime := now.Add(3 * time.Minute)
		assert.Equal(t,records[1:],FilterRecordsBefore(records,checkTime))
	})
	t.Run("only 1 record",func(t *testing.T){
		records := []Record{{Time:now,  Usage: 1e9}}
		assert.Equal(t,records,FilterRecordsBefore(records,now))
	})
	t.Run("when no records after migration time, just return previous record with migration time",func(t *testing.T){
		checkTime := now.Add(4 * time.Minute)
		records := []Record{{Time:now,  Usage: 1e9}}
		assert.Equal(t,[]Record{{Time:checkTime,  Usage: 1e9}},FilterRecordsBefore(records,checkTime))
	})
}

func TestSetPodResources(t *testing.T) {
	t.Run("get S size", func(t *testing.T){
		resourceS :=v1.ResourceList{
		  "cpu":            resource.MustParse("5"),
		  "memory":         resource.MustParse("30Gi"),
		}
		assert.Equal(t,resourceS,GetJobResourceRequest("s"))
	})
	t.Run("get M size", func(t *testing.T) {
		resourceM :=v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("80Gi"),
		      }
		assert.Equal(t,resourceM,GetJobResourceRequest("m"))	
	})
	t.Run("get L size", func(t *testing.T) {
		resource :=v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("130Gi"),
		      }
		assert.Equal(t,resource,GetJobResourceRequest("l"))	
	})
	t.Run("get XL size", func(t *testing.T) {
		resource :=v1.ResourceList{
			"cpu":            resource.MustParse("8"),
			"memory":         resource.MustParse("420Gi"),
		      }
		assert.Equal(t,resource,GetJobResourceRequest("xl"))	
	})


    }


