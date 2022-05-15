package migration

type Client struct {
	UsedMemory int64
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) UpdateMetrics(memory int64) {
	c.UsedMemory = memory
}
