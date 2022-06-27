package node_test

// func TestTaintNode(t *testing.T) {
// 	stime := "2022-05-11T08:00:00+02:00"
// 	nv1, err := config.BuildNode(config.NodeConfig{Metadata:metav1.ObjectMeta{Name: "zone"},Status:config.NodeStatus{Allocatable: map[v1.ResourceName]string{"cpu": "120", "memory": "450G","pods":"1000"}}},stime)
// 	assert.NoError(t,err)
// 	sut := node.NewNode(nv1)
// 	// ctime, err := time.Parse(time.RFC3339, stime)
// 	// podmem := jobparser.PodMemory{Name: "o10n-worker-m-zx8wp-n5", Records: []jobparser.Record{{Time: time.Now(), Usage: 1e9}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e2}}}
// 	// pod := jobparser.CreatePodWithoutResources(podmem)

// 	sut.Unschedulable()
// 	assert.Equal(t,"",sut)
// 	// _,err = sut.BindPod(clock.NewClock(ctime),pod)
// 	assert.NoError(t, err)
// }
