package migration

import (
	"github.com/containerd/containerd/log"
	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

type Unscheduler struct {
	EndTime clock.Clock
	ThresholdDecimal float64
	ReschedulableDistanceDecimal float64 // condition to set unschedulable node to schedulable when decimal distance to the threshold is bigger or equal to ReschedulableDistanceDecimal
}

func (unsched *Unscheduler) Submit(
	currentTime clock.Clock,
	nodeLister algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error){
		nodes,_ := nodeLister.List()
		for name,node := range met[metrics.NodesMetricsKey].(map[string]node.Metrics) {
			usage :=  float64(node.TotalResourceUsage.Memory().ScaledValue(resource.Giga))
			alloc := float64(node.Allocatable.Memory().ScaledValue(resource.Giga))
			usedDecimal := usage / alloc
			if usedDecimal> unsched.ThresholdDecimal {
				if res := GetNodeWithName(name,nodes); res != nil {
					if !res.Spec.Unschedulable {
						res.Spec.Unschedulable = true
						log.L.Debugf("Node %s  (used decimal: %d ) is set to unschedulable",name,usedDecimal)
					}
				}			
			} else if usedDecimal<= unsched.ThresholdDecimal - unsched.ReschedulableDistanceDecimal {
				if res := GetNodeWithName(name,nodes); res != nil {
					if res.Spec.Unschedulable {
						log.L.Debugf("Node %s (used decimal: %d ) is set to schedulable again",name,usedDecimal)
						res.Spec.Unschedulable = false
					}
				}
			}
		}


		isSimulationFinished :=unsched.EndTime.BeforeOrEqual(currentTime)
		if isSimulationFinished {
			return []submitter.Event{&submitter.TerminateSubmitterEvent{}},nil
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

