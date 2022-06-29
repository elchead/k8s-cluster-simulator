package migration

import (
	"errors"

	"github.com/elchead/k8s-cluster-simulator/pkg/pod"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
)


type Memorizer[T interface{}] struct {
	MemoInterval int
	data T
	prior T
	currentStep int
}

func (m *Memorizer[T]) Update(data T) {
	m.data = data
	if m.currentStep % m.MemoInterval == 0 {
		m.prior = m.data
	}
	m.currentStep++
}

func (m *Memorizer[T]) Value() T {
	return m.data
}

func (m *Memorizer[T]) Prior() T {
	return m.prior
}

type Client struct {
	UsedMemoryMap map[string]int64 // key: nodeName
	TotalMemoryMap map[string]int64 // key: nodeName
	PodMemoryMap map[string]monitoring.PodMemMap // key: nodeName
	PodMemorizer map[string]*Memorizer[monitoring.PodMemMap]
	MemoInterval int
}

func NewClient() *Client {
	return NewClientWithMemoStep(5)
}

func NewClientWithMemoStep(memostep int) *Client {
	return &Client{UsedMemoryMap: make(map[string]int64), TotalMemoryMap: make(map[string]int64), PodMemoryMap: make(map[string]monitoring.PodMemMap),PodMemorizer: make(map[string]*Memorizer[monitoring.PodMemMap]),MemoInterval: memostep}
}

func (c *Client) UpdatePodMetric(podname string,pd pod.Metrics) {
	intUsage :=  pd.ResourceUsage.Memory().ScaledValue(resource.Giga)
	if len(c.PodMemoryMap[pd.Node]) == 0 {
		c.PodMemoryMap[pd.Node] = make(monitoring.PodMemMap)		
	} 
	c.PodMemoryMap[pd.Node][podname] = float64(intUsage)
}

func (c *Client) UpdatePodMetrics(pods map[string]pod.Metrics) {
	c.PodMemoryMap = make(map[string]monitoring.PodMemMap)
	for podname,pod := range pods {
		c.UpdatePodMetric(podname,pod)
	}
	c.updateMemorizer()
}

func (c *Client) updateMemorizer() {
	for node, pod := range c.PodMemoryMap {
		if c.PodMemorizer[node] == nil {
			c.PodMemorizer[node] = &Memorizer[monitoring.PodMemMap]{MemoInterval: c.MemoInterval}
		}
		c.PodMemorizer[node].Update(pod.Copy())
	}
}

func (c Client) GetPodMemorySlope(node, name, time, slopeWindow string) (float64, error) {
	val,ok := c.PodMemorizer[node].Value()[name]
	pval,pok := c.PodMemorizer[node].Prior()[name]
	if !ok || !pok {
		return -1., errors.New("could not get pod memory for node " +name)
	}
	return val - pval,nil
}

// in Gb
func (c *Client) GetPodMemories(node string) (monitoring.PodMemMap,error) {
	val, ok := c.PodMemoryMap[node]
	if !ok {
		return nil, errors.New("could not get pod memory for node " +node)
	}
	// log.L.Debug("Podmemories",name, val)
	return val,nil
}

func (c *Client) UpdateNodeMetrics(metrics map[string]node.Metrics) {
	c.UsedMemoryMap = make(map[string]int64)
	c.TotalMemoryMap = make(map[string]int64)
	for node, metric := range metrics {
		c.UsedMemoryMap[node] = metric.TotalResourceUsage.Memory().ScaledValue(resource.Giga)
		c.TotalMemoryMap[node] = metric.Allocatable.Memory().ScaledValue(resource.Giga)
	}
}

func (c *Client) GetFreeMemoryNode(name string) (float64, error) {
	usedGb, ok := c.UsedMemoryMap[name]
	if !ok {
		return 0, errors.New("could not get free memory for node " +name)
	}
	total, ok := c.TotalMemoryMap[name]
	if !ok {
		return 0, errors.New("could not get total memory for node " +name)
	}
	return 100.- float64(usedGb)/float64(total)*100., nil
}

func (c *Client) 	GetFreeMemoryOfNodes() (monitoring.NodeFreeMemMap, error) {
	res := make(monitoring.NodeFreeMemMap)
	for node, _ := range c.UsedMemoryMap {
		free, err := c.GetFreeMemoryNode(node)
		if err != nil {
			return nil, err
		}
		res[node] = free
	}
	// log.L.Debug("Nodememory",res )

	return res, nil
}


