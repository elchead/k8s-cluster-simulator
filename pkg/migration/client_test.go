package migration_test

import (
	"testing"

	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/config"
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
	
	t.Run("update node metrics", func(t *testing.T) {
		nodeSz, err := resource.ParseQuantity("450Gi")
		assert.NoError(t, err)
		usageSz, _ := resource.ParseQuantity("5Gi")

		metrics := node.Metrics{Allocatable: v1.ResourceList{"memory": nodeSz,},TotalResourceUsage:v1.ResourceList{"memory": usageSz}} //metrics.Metrics{}
		sut.UpdateNodeMetrics(metrics)
		assert.Equal(t,int64(5368709120),sut.UsedMemory)
		assert.Equal(t,int64(483183820800),sut.TotalMemory)
		
	})
}

func newNode(name, memCapacity string) (node.Node,error) {
	resources := map[v1.ResourceName]string{"memory": memCapacity}

	nodeConfig := config.NodeConfig{Metadata:  metav1.ObjectMeta{Name:"zone2"},Spec: v1.NodeSpec{},Status:config.NodeStatus{Allocatable: resources}} 

	v1Node, err := config.BuildNode(nodeConfig,"2022-05-11T08:00:00Z")

	return node.NewNode(v1Node),err
}
