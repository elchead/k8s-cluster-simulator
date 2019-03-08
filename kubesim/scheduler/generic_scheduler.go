package scheduler

import (
	"errors"
	"fmt"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
	"github.com/ordovicia/kubernetes-simulator/log"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// GenericScheduler makes scheduling decision for each given pod in the one-by-one manner.
// This type is similar to "k8s.io/pkg/Scheduler/Scheduler/core".genericScheduler, which implements
// "k8s.io/pkg/Scheduler/Scheduler/core".ScheduleAlgorithm.
type GenericScheduler struct {
	extenders []Extender

	predicates   map[string]predicates.FitPredicate
	prioritizers []priorities.PriorityConfig

	pendingPods   []*v1.Pod
	lastNodeIndex uint64
}

// NewGenericScheduler creates a new GenericScheduler.
func NewGenericScheduler() GenericScheduler {
	return GenericScheduler{
		predicates: map[string]predicates.FitPredicate{},
	}
}

// AddExtender adds an extender to this GenericScheduler.
func (sched *GenericScheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// AddPredicate adds a predicate plugin to this GenericScheduler.
func (sched *GenericScheduler) AddPredicate(name string, predicate predicates.FitPredicate) {
	sched.predicates[name] = predicate
}

// AddPrioritizer adds a prioritizer plugin to this GenericScheduler.
func (sched *GenericScheduler) AddPrioritizer(prioritizer priorities.PriorityConfig) {
	sched.prioritizers = append(sched.prioritizers, prioritizer)
}

// Schedule implements Scheduler interface.
func (sched *GenericScheduler) Schedule(
	clock clock.Clock,
	podQueue queue.PodQueue,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error) {

	results := []Event{}

	for {
		pod, err := podQueue.Front()
		if err != nil {
			if err == queue.ErrEmptyQueue {
				break
			} else {
				return results, errors.New("Unexpected error raised by Queueu.Pop()")
			}
		}

		log.L.Tracef("Trying to schedule pod %v", pod)

		podKey, err := util.PodKey(pod)
		if err != nil {
			return results, err
		}
		log.L.Debugf("Trying to schedule pod %s", podKey)

		result, err := sched.scheduleOne(pod, nodeLister, nodeInfoMap)
		if err != nil {
			updatePodStatusSchedulingFailure(clock, pod, err)

			if _, ok := err.(*core.FitError); ok {
				log.L.Tracef("Pod %v does not fit in any node", pod)
				log.L.Debugf("Pod %s does not fit in any node", podKey)
				break
			} else {
				return []Event{}, nil
			}
		}

		log.L.Debugf("Selected node %s", result.SuggestedHost)

		pod, _ = podQueue.Pop()
		updatePodStatusSchedulingSucceess(clock, pod)

		nodeInfo, ok := nodeInfoMap[result.SuggestedHost]
		if !ok {
			return []Event{}, fmt.Errorf("No node named %s", result.SuggestedHost)
		}
		nodeInfo.AddPod(pod)

		results = append(results, &BindEvent{Pod: pod, ScheduleResult: result})
	}

	return results, nil
}

// func (sched *GenericScheduler) Preempt(
// 	pod *v1.Pod,
// 	nodeLister algorithm.NodeLister,
// 	err error) (selectedNode *v1.Node,
// 	preemptedPods []*v1.Pod,
// 	cleanupNominatedPods []*v1.Pod, e error) {
// }

var _ = Scheduler(&GenericScheduler{})

// scheduleOne makes scheduling decision for the given pod and nodes.
// Returns core.ErrNoNodesAvailable if nodeLister lists zero nodes, or core.FitError if the given
// pod does not fit in any nodes.
func (sched *GenericScheduler) scheduleOne(
	pod *v1.Pod,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) (core.ScheduleResult, error) {

	result := core.ScheduleResult{}
	nodes, err := nodeLister.List()
	if err != nil {
		return result, err
	}
	if len(nodes) == 0 {
		return result, core.ErrNoNodesAvailable
	}

	nodesFiltered, failedPredicateMap, err := sched.filter(pod, nodes, nodeInfoMap)
	if err != nil {
		return result, err
	}

	switch len(nodesFiltered) {
	case 0:
		return result, &core.FitError{
			Pod:              pod,
			NumAllNodes:      len(nodes),
			FailedPredicates: failedPredicateMap,
		}
	case 1:
		return core.ScheduleResult{
			SuggestedHost:  nodesFiltered[0].Name,
			EvaluatedNodes: 1 + len(failedPredicateMap),
			FeasibleNodes:  1,
		}, nil
	}

	prios, err := sched.prioritize(pod, nodesFiltered, nodeInfoMap)
	if err != nil {
		return result, err
	}
	host, err := sched.selectHost(prios)

	return core.ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(nodesFiltered) + len(failedPredicateMap),
		FeasibleNodes:  len(nodesFiltered),
	}, err
}

func (sched *GenericScheduler) filter(
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]*v1.Node, core.FailedPredicateMap, error) {

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Filtering nodes %v", nodeNames)

	failedPredicateMap := core.FailedPredicateMap{}
	filteredNodes := nodes

	// In-process plugins
	errs := kerr.MessageCountMap{}
	for name, p := range sched.predicates {
		var err error
		filteredNodes, err = callPredicatePlugin(name, &p, pod, filteredNodes, nodeInfoMap, failedPredicateMap, errs)
		if err != nil {
			return []*v1.Node{}, core.FailedPredicateMap{}, err
		}

		if len(filteredNodes) == 0 {
			break
		}
	}

	if len(errs) > 0 {
		return []*v1.Node{}, core.FailedPredicateMap{}, kerr.CreateAggregateFromMessageCountMap(errs)
	}

	// Extenders
	if len(filteredNodes) > 0 && len(sched.extenders) > 0 {
		for _, extender := range sched.extenders {
			var err error
			filteredNodes, err = extender.filter(pod, filteredNodes, nodeInfoMap, failedPredicateMap)
			if err != nil {
				return []*v1.Node{}, core.FailedPredicateMap{}, err
			}

			if len(filteredNodes) == 0 {
				break
			}
		}
	}

	nodeNames = make([]string, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Filtered nodes %v", nodeNames)

	return filteredNodes, failedPredicateMap, nil
}

func (sched *GenericScheduler) prioritize(
	pod *v1.Pod,
	filteredNodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) (api.HostPriorityList, error) {

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Prioritizing nodes %v", nodeNames)

	prioList := make(api.HostPriorityList, len(filteredNodes), len(filteredNodes))

	// If no priority configs are provided, then the EqualPriority function is applied
	// This is required to generate the priority list in the required format
	if len(sched.prioritizers) == 0 && len(sched.extenders) == 0 {
		for i, node := range filteredNodes {
			nodeInfo, ok := nodeInfoMap[node.Name]
			if !ok {
				return api.HostPriorityList{}, fmt.Errorf("No node named %s", node.Name)
			}

			prio, err := core.EqualPriorityMap(pod, &dummyPriorityMetadata{}, nodeInfo)
			if err != nil {
				return api.HostPriorityList{}, err
			}
			prioList[i] = prio
		}
		return prioList, nil
	}

	for i, node := range filteredNodes {
		prioList[i] = api.HostPriority{Host: node.Name, Score: 0}
	}

	errs := []error{}

	// In-process plugins
	for _, prioritizer := range sched.prioritizers {
		prios, err := callPrioritizePlugin(&prioritizer, pod, filteredNodes, nodeInfoMap, errs)
		if err != nil {
			return api.HostPriorityList{}, err
		}

		for i, prio := range prios {
			prioList[i].Score += prio.Score
		}
	}

	if len(errs) > 0 {
		return api.HostPriorityList{}, kerr.NewAggregate(errs)
	}

	// Extenders
	if len(sched.extenders) > 0 {
		prioMap := map[string]int{}
		for _, extender := range sched.extenders {
			extender.prioritize(pod, filteredNodes, prioMap)
		}

		for i, prio := range prioList {
			prioList[i].Score += prioMap[prio.Host]
		}
	}

	log.L.Debugf("Prioritized nodes %v", prioList)

	return prioList, nil
}

// selectHost is copied from "k8s.io/kubernetes/pkg/scheduler/core".selectHost().
func (sched *GenericScheduler) selectHost(priorities api.HostPriorityList) (string, error) {
	if len(priorities) == 0 {
		return "", errors.New("Empty priorities")
	}

	maxScores := findMaxScores(priorities)
	idx := int(sched.lastNodeIndex % uint64(len(maxScores)))
	sched.lastNodeIndex++

	return priorities[maxScores[idx]].Host, nil
}

// findMaxScores is copied from "k8s.io/kubernetes/pkg/scheduler/core".findMaxScores().
func findMaxScores(priorities api.HostPriorityList) []int {
	maxScoreIndexes := make([]int, 0, len(priorities)/2)
	maxScore := priorities[0].Score

	for idx, prio := range priorities {
		if prio.Score > maxScore {
			maxScore = prio.Score
			maxScoreIndexes = maxScoreIndexes[:0]
			maxScoreIndexes = append(maxScoreIndexes, idx)
		} else if prio.Score == maxScore {
			maxScoreIndexes = append(maxScoreIndexes, idx)
		}
	}

	return maxScoreIndexes
}

func updatePodStatusSchedulingSucceess(clock clock.Clock, pod *v1.Pod) {
	util.UpdatePodCondition(clock, &pod.Status, &v1.PodCondition{
		Type:          v1.PodScheduled,
		Status:        v1.ConditionTrue,
		LastProbeTime: clock.ToMetaV1(),
		// Reason:
		// Message:
	})
}

func updatePodStatusSchedulingFailure(clock clock.Clock, pod *v1.Pod, err error) {
	util.UpdatePodCondition(clock, &pod.Status, &v1.PodCondition{
		Type:          v1.PodScheduled,
		Status:        v1.ConditionFalse,
		LastProbeTime: clock.ToMetaV1(),
		Reason:        v1.PodReasonUnschedulable,
		Message:       err.Error(),
	})
}