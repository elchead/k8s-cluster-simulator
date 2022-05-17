package migration

import (
	"errors"

	"github.com/containerd/containerd/log"
	"github.com/elchead/k8s-cluster-simulator/pkg/pod"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
)
type Client struct {
	UsedMemoryMap map[string]	int64 // key: nodeName
	TotalMemoryMap map[string]int64 // key: nodeName
	PodMemoryMap map[string]monitoring.PodMemMap // key: nodeName
}

func NewClient() *Client {
	return &Client{UsedMemoryMap: make(map[string]int64), TotalMemoryMap: make(map[string]int64), PodMemoryMap: make(map[string]monitoring.PodMemMap)}
}

func (c *Client) UpdatePodMetric(podname string,pd pod.Metrics) {
	intUsage :=  pd.ResourceUsage.Memory().ScaledValue(resource.Giga)
	if len(c.PodMemoryMap[pd.Node]) == 0 {
		c.PodMemoryMap[pd.Node] = make(monitoring.PodMemMap)		
	} 
	podmap := c.PodMemoryMap[pd.Node]
	podmap[podname] = float64(intUsage)
}

func (c *Client) UpdatePodMetrics(pods map[string]pod.Metrics) {
	for podname,pod := range pods {
		c.UpdatePodMetric(podname,pod)
	}
}

// in Gb
func (c *Client) GetPodMemories(name string) (monitoring.PodMemMap,error) {
	val, ok := c.PodMemoryMap[name]
	if !ok {
		return nil, errors.New("could not get pod memory for node " +name)
	}
	log.L.Debug("Podmemories",name, val)
	return val,nil
}

func (c *Client) UpdateNodeMetrics(metrics map[string]node.Metrics) {
	for node, metric := range metrics {
		c.UsedMemoryMap[node] = metric.TotalResourceUsage.Memory().ScaledValue(resource.Giga)
		c.TotalMemoryMap[node] = metric.Allocatable.Memory().ScaledValue(resource.Giga)
	}
}

func (c *Client) GetFreeMemoryNode(name string) (float64, error) {
	free, ok := c.UsedMemoryMap[name]
	if !ok {
		return 0, errors.New("could not get free memory for node " +name)
	}
	total, ok := c.TotalMemoryMap[name]
	if !ok {
		return 0, errors.New("could not get total memory for node " +name)
	}
	return 100.- float64(free)/float64(total)*100., nil
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


