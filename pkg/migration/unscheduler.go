package migration

import (
	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

type Unscheduler struct {}

func (unsched *Unscheduler) Submit(
	clock clock.Clock,
	nodeLister algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error){
		nodes,_ := nodeLister.List()
		// if node metric usage percentage is >80 -> set node to unschedulable
		for name,node := range met[metrics.NodesMetricsKey].(map[string]node.Metrics) {
			usage := node.TotalResourceUsage.Memory().ScaledValue(resource.Giga)
			alloc := node.Allocatable.Memory().ScaledValue(resource.Giga)
			if float64(usage) / float64(alloc) > .8 {
				if res := GetNodeWithName(name,nodes); res != nil {
					res.Spec.Unschedulable = true
				}			
			}
		}
		return nil, nil
	}

func GetNodeWithName(name string, nodes []*v1.Node) *v1.Node {
	for _, node := range nodes {
		if node.ObjectMeta.Name == name {
			return node
		}
	}
	return nil
}

