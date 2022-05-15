package migration

import (
	"errors"

	"github.com/elchead/k8s-cluster-simulator/pkg/node"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
)
type Client struct {
	UsedMemoryMap map[string]	int64
	TotalMemoryMap map[string]int64
	PodMemory float64
}

func NewClient() *Client {
	return &Client{UsedMemoryMap: make(map[string]int64), TotalMemoryMap: make(map[string]int64)}
}

func (c *Client) UpdatePodMemory(value float64) {
	c.PodMemory =  value
}

func (c *Client) GetPodMemories(name string) (monitoring.PodMemMap,error) {
	return monitoring.PodMemMap{"pod":c.PodMemory},nil
}

func (c *Client) UpdateNodeMetrics(metrics map[string]node.Metrics) {
	for node, metric := range metrics {
		c.UsedMemoryMap[node], _ = metric.TotalResourceUsage.Memory().AsInt64()
		c.TotalMemoryMap[node], _ = metric.Allocatable.Memory().AsInt64()
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
	return res, nil
}


