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
func TestCreateNode(t *testing.T) {
	node,err := newNode("zone2","450Gi")
	assert.NoError(t, err)
	nodeInfo, _ := node.ToNodeInfo(clock.NewClock(time.Now()))
	assert.Equal(t,int64(483183820800),nodeInfo.AllocatableResource().Memory)	
}

func TestUpdateMetrics(t *testing.T) {
	sut := migration.NewClient()
	
	t.Run("get free memory percentage of specfic node",func(t *testing.T) {
		nodemetric := createNodeMetrics(500,5)
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
		metrics := map[string]node.Metrics{"zone2": createNodeMetrics(500,5),"zone3": createNodeMetrics(500,10)}
		sut.UpdateNodeMetrics(metrics)
		free,err :=sut.GetFreeMemoryOfNodes()
		assert.NoError(t, err)
		assert.Equal(t,monitoring.NodeFreeMemMap{"zone2":99.,"zone3":98.},free)

	})
	t.Run("delete old metrics upon update", func(t *testing.T) {
		metrics := map[string]node.Metrics{"zone3": createNodeMetrics(400,10)}
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

func createNodeMetrics(total,used int64) node.Metrics {
	return node.Metrics{Allocatable: createMemoryResource(total),TotalResourceUsage:createMemoryResource(used)}

}

func createMemoryResource(quantity int64) v1.ResourceList {
	return v1.ResourceList{"memory": *resource.NewScaledQuantity(quantity,resource.Giga),}	
}

func newNode(name, memCapacity string) (node.Node,error) {
	resources := map[v1.ResourceName]string{"memory": memCapacity}

	nodeConfig := config.NodeConfig{Metadata:  metav1.ObjectMeta{Name:"zone2"},Spec: v1.NodeSpec{},Status:config.NodeStatus{Allocatable: resources}} 

	v1Node, err := config.BuildNode(nodeConfig,"2022-05-11T08:00:00Z")

	return node.NewNode(v1Node),err
}
