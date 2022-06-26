package migration_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnscheduleWhenNodeFull(t *testing.T) {
	c := clock.NewClock(time.Now())
	sut := migration.Unscheduler{EndTime:c.Add(5* time.Minute),ThresholdDecimal: .8,ReschedulableDistanceDecimal:.15}
	
	nodes := []*v1.Node{{ObjectMeta:metav1.ObjectMeta{Name:"zone3"},Spec:  v1.NodeSpec{Unschedulable: false}},{ObjectMeta:metav1.ObjectMeta{Name:"zone2"},Spec:  v1.NodeSpec{Unschedulable: false}}}
	fakeSim := FakeSimulator{nodes}

	met := createNodeMetrics(100.,map[string]int64{"zone3":10,"zone2":85})
	
	sut.Submit(c,fakeSim,met)
	t.Run("node with usage < 80% remains schedulable", func(t *testing.T) {
		assert.False(t,migration.GetNodeWithName("zone3",nodes).Spec.Unschedulable)
	})
	t.Run("node zone2 with usage > 80% is set unschedulable",func(t *testing.T) {
		assert.True(t,migration.GetNodeWithName("zone2",nodes).Spec.Unschedulable)
	})
	t.Run("zone2 is set to schedulable again when 15 % away from threshold",func(t *testing.T) {
		met = createNodeMetrics(100.,map[string]int64{"zone3":10,"zone2":64})
		sut.Submit(c.Add(5* time.Minute),fakeSim,met)
		assert.False(t,migration.GetNodeWithName("zone2",nodes).Spec.Unschedulable)
	})

}

func createNodeMetrics(nodeCapacity int64, nodeUsage map[string]int64) map[string]interface{} {
	met := make(map[string]interface{})
	nodesMetrics := make(map[string]node.Metrics)
	for name,usage := range nodeUsage {
		nodesMetrics[name] = node.Metrics{
			Allocatable:        createMemoryResource(nodeCapacity),
			TotalResourceUsage: createMemoryResource(usage),
		}
	}
	met[metrics.NodesMetricsKey] = nodesMetrics
	return met
}

type FakeSimulator struct {
	nodes []*v1.Node
}

func (s FakeSimulator) List()  ([]*v1.Node, error) {
	return s.nodes,nil	
}
