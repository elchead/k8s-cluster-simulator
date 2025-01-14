package migration_test

import (
	"testing"

	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/config"
	"github.com/elchead/k8s-cluster-simulator/pkg/pod"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
	"k8s.io/apimachinery/pkg/api/resource"

	// "github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClientReturnsRuntimeAndExecutedSeconds(t *testing.T) {
	sut := migration.NewClient()	
	podmetrics := map[string]pod.Metrics{"worker":{Runtime:3600,ExecutedSeconds:1800,Node:"zone3",ResourceUsage:createMemoryResource(50.)}}
	sut.UpdatePodMetrics(podmetrics)
	assert.Equal(t,int32(3600),sut.GetRuntime("worker"))
	assert.Equal(t,int32(1800),sut.GetExecutionTime("worker"))
}

func TestMemorizer(t *testing.T){
	sut := &migration.Memorizer[int]{MemoInterval: 3}
	sut.Update(1)
	t.Run("return default when no prior available yet", func(t *testing.T){
		assert.Empty(t,sut.Prior())
	})
	sut.Update(2)
	sut.Update(3)
	assert.Equal(t,3,sut.Value())
	assert.Equal(t,1,sut.Prior())
	t.Run("update slope after each step",func(t *testing.T){
		sut.Update(4)
		assert.Equal(t,4,sut.Value())
		assert.Equal(t,2,sut.Prior())
	})
	sut.Update(5)
	sut.Update(6)
	assert.Equal(t,4,sut.Prior())
	
}

func TestMemorizerWithPodMemMap(t *testing.T){
	m := make(map[string]monitoring.PodMemMap)
	m["z1"] = make(monitoring.PodMemMap)
	m["z1"]["worker"] = 1
	sut := &migration.Memorizer[monitoring.PodMemMap]{MemoInterval: 3}
	// assert.Equal(t,"",sut.Value())
	sut.Update(m["z1"].Copy())
	m["z1"]["worker"] = 2
	sut.Update(m["z1"].Copy())
	m["z1"]["worker"] = 3
	sut.Update(m["z1"].Copy())
	assert.Equal(t,1.,sut.Prior()["worker"])
}

func TestGetSlope(t *testing.T) {
	sut := migration.NewClientWithMemoStep(3)

	t.Run("no slope when no prior data available", func(t *testing.T){
		noRes, err := sut.GetPodMemorySlope("zone2","worker2","now","1m")
		assert.Error(t, err)
		assert.Equal(t,-1.,noRes)
	})
	podmetrics := map[string]pod.Metrics{"worker":{Node:"zone3",ResourceUsage:createMemoryResource(50.)}}
	sut.UpdatePodMetrics(podmetrics)
	podmetrics = map[string]pod.Metrics{"worker":{Node:"zone3",ResourceUsage:createMemoryResource(75.)}}
	sut.UpdatePodMetrics(podmetrics)
	podmetrics = map[string]pod.Metrics{"worker":{Node:"zone3",ResourceUsage:createMemoryResource(80.)}}
	sut.UpdatePodMetrics(podmetrics)

	t.Run("get slope on different node", func(t *testing.T){
		podmetrics := map[string]pod.Metrics{"worker2":{Node:"zone2",ResourceUsage:createMemoryResource(10.)},"worker3":{Node:"zone2",ResourceUsage:createMemoryResource(20.)}}
		sut.UpdatePodMetrics(podmetrics)
		podmetrics = map[string]pod.Metrics{"worker2":{Node:"zone2",ResourceUsage:createMemoryResource(75.)},"worker3":{Node:"zone2",ResourceUsage:createMemoryResource(20.)}}
		sut.UpdatePodMetrics(podmetrics)
		podmetrics = map[string]pod.Metrics{"worker2":{Node:"zone2",ResourceUsage:createMemoryResource(30.)},"worker3":{Node:"zone2",ResourceUsage:createMemoryResource(30.)}}
		sut.UpdatePodMetrics(podmetrics)

		res2, err := sut.GetPodMemorySlope("zone2","worker2","now","1m")
		assert.NoError(t, err)
		assert.Equal(t,20.,res2)

		res3, err := sut.GetPodMemorySlope("zone2","worker3","now","1m")
		assert.NoError(t, err)
		assert.Equal(t,10.,res3)
	})

	res, err := sut.GetPodMemorySlope("zone3","worker","now","1m")
	assert.NoError(t, err)
	assert.Equal(t,30.,res)
	
}
func TestCreateNode(t *testing.T) {
	node,err := newNode("zone2","450Gi")
	assert.NoError(t, err)
	nodeInfo, _ := node.ToNodeInfo(clock.NewClock(time.Now()))
	assert.Equal(t,int64(483183820800),nodeInfo.AllocatableResource().Memory)	
}

func TestUpdateMetrics(t *testing.T) {
	sut := migration.NewClient()
	
	t.Run("get free memory percentage of specfic node",func(t *testing.T) {
		nodemetric := createNodeMetric(500,5)
		metrics := map[string]node.Metrics{"zone2": nodemetric}
		sut.UpdateNodeMetrics(metrics)
		free,err :=sut.GetFreeMemoryNode("zone2")
		assert.NoError(t, err)
		assert.Equal(t,99.,free)	

		t.Run("fail for non existing node", func(t *testing.T) {
			_,err :=sut.GetFreeMemoryNode("zone3")
			assert.Error(t, err)

		})
	})
	t.Run("get free memory of all nodes", func(t *testing.T) {
		metrics := map[string]node.Metrics{"zone2": createNodeMetric(500,5),"zone3": createNodeMetric(500,10)}
		sut.UpdateNodeMetrics(metrics)
		free,err :=sut.GetFreeMemoryOfNodes()
		assert.NoError(t, err)
		assert.Equal(t,monitoring.NodeFreeMemMap{"zone2":99.,"zone3":98.},free)

	})
	t.Run("delete old metrics upon update", func(t *testing.T) {
		metrics := map[string]node.Metrics{"zone3": createNodeMetric(400,10)}
		sut.UpdateNodeMetrics(metrics)
		free,_ :=sut.GetFreeMemoryOfNodes()
		assert.Equal(t,monitoring.NodeFreeMemMap{"zone3":97.5},free)
	})
	}

func TestGetPodMemories(t *testing.T) {
	sut := migration.NewClient()
	t.Run("fail to get free pod memories if not existent for node", func(t *testing.T) {
		podmetrics := pod.Metrics{Node:"zone3",ResourceUsage:createMemoryResource(50.)}
		sut.UpdatePodMetric("worker",podmetrics)
		_, err := sut.GetPodMemories("zone1")
		assert.Error(t, err)
	})
	t.Run("get free pod memories if node existent", func(t *testing.T) {
		podmetrics := pod.Metrics{Node:"zone3",ResourceUsage:createMemoryResource(50)}
		sut.UpdatePodMetric("worker",podmetrics)
		res, err := sut.GetPodMemories("zone3")
		assert.NoError(t, err)
		assert.Equal(t, 50.,res["worker"])
	})
	t.Run("update multiple pod metrics and get free pod memories", func(t *testing.T) {
		workerMetrics := pod.Metrics{Node:"zone3",ResourceUsage:createMemoryResource(50)}
		z2Metrics := pod.Metrics{Node:"zone2",ResourceUsage:createMemoryResource(50)}

		sut.UpdatePodMetrics(map[string]pod.Metrics{"z3OLD_worker":workerMetrics,"z2_worker":z2Metrics})		
		sut.UpdatePodMetrics(map[string]pod.Metrics{"z3_worker":workerMetrics,"z2_worker":z2Metrics})

		res, err := sut.GetPodMemories("zone3")
		assert.NoError(t, err)
		t.Run("do not get pod from other node", func(t *testing.T) {
			_,ok := res["z2_worker"]
			assert.False(t,ok)
		})
		t.Run("get pod from node", func(t *testing.T) {
			mem,ok := res["z3_worker"]
			assert.True(t,ok)
			assert.Equal(t, 50.,mem)
		})
		t.Run("delete old metrics upon update", func(t *testing.T) {
			_,ok := res["z3OLD_worker"]
			assert.False(t,ok)
		})
	})

}

func createNodeMetric(total,used int64) node.Metrics {
	return node.Metrics{Allocatable: createMemoryResource(total),TotalResourceUsage:createMemoryResource(used)}

}

func createMemoryResource(quantityGb int64) v1.ResourceList {
	return v1.ResourceList{"memory": *resource.NewScaledQuantity(quantityGb,resource.Giga),}	
}

func newNode(name, memCapacity string) (node.Node,error) {
	resources := map[v1.ResourceName]string{"memory": memCapacity}

	nodeConfig := config.NodeConfig{Metadata:  metav1.ObjectMeta{Name:"zone2"},Spec: v1.NodeSpec{},Status:config.NodeStatus{Allocatable: resources}} 

	v1Node, err := config.BuildNode(nodeConfig,"2022-05-11T08:00:00Z")

	return node.NewNode(v1Node),err
}
