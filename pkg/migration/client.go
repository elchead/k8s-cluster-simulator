package migration

import (
	"github.com/elchead/k8s-cluster-simulator/pkg/node"
)
type Client struct {
	UsedMemory int64
	TotalMemory int64
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) UpdateNodeMetrics(metrics node.Metrics) {
	c.UsedMemory, _= metrics.TotalResourceUsage.Memory().AsInt64()
	c.TotalMemory, _ = metrics.Allocatable.Memory().AsInt64()
}

func (c *Client) GetFreeMemoryNode(name string) (float64, error) {
	return 100.- float64(c.UsedMemory)/float64(c.TotalMemory)*100., nil
}


