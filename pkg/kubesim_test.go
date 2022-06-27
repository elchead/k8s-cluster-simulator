// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubesim_test

import (
	"context"
	"testing"
	"time"

	kubesim "github.com/elchead/k8s-cluster-simulator/pkg"
	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/queue"
	"github.com/elchead/k8s-cluster-simulator/pkg/scheduler"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

func TestSimFreeze(t *testing.T) {
	queue := queue.NewPriorityQueue()
	sched := scheduler.NewGenericScheduler( /* preemption enabled */ true)
	metricClient := migration.NewClient()

	configPath := "config"
	conf, err := kubesim.ReadConfig(configPath)
	assert.NoError(t, err)
	// if useMigrator {
	// 	for i, logger := range conf.MetricsLogger {
	// 		conf.MetricsLogger[i].Dest = "m-" + logger.Dest
	// 	}
	// }
	// if err != nil {
	// 	log.L.Fatal("Failed to read config:", err)
	// }
	sim,_ := kubesim.NewKubeSim(conf, queue, &sched,metricClient)

	f := jobparser.PodFactory{SetResources: false}
	now,err := time.Parse(time.RFC3339, conf.StartClock)//time.Now()
	assert.NoError(t, err)
	later := now.Add(2 * time.Minute)
	podmem := jobparser.PodMemory{Name: "po", Records: []jobparser.Record{{Time: now, Usage: 1e12}, {Time: later, Usage: 1e9},{Time: later.Add(2 * time.Minute), Usage: 1},{Time: later.Add(30 * time.Minute), Usage: 1e6}}} // latest spec entry denotes termination time
	pod := f.New(podmem)
	s := &TestSubmitter{}
	s.On("Submit", mock.Anything, mock.Anything, mock.Anything).Return([]submitter.Event{&submitter.SubmitEvent{Pod:pod}}, nil).Once()
	s.On("Submit", mock.Anything, mock.Anything,mock.Anything).Return([]submitter.Event{&submitter.FreezeUsageEvent{PodKey:"default/po"}}, nil).Once()
	s.On("Submit", mock.Anything, mock.Anything,mock.Anything).Return([]submitter.Event{&submitter.TerminateSubmitterEvent{}}, nil)
	sim.AddSubmitter("TestSubmitter", s)
	sim.Run(context.TODO())
	// sim.BuildMetrics()
	// assert.True(t,false)
}

type TestSubmitter struct {
	Events []submitter.Event
	mock.Mock
}

func (t *TestSubmitter) Submit(
	clock clock.Clock,
	nodeLister algorithm.NodeLister,
	metrics metrics.Metrics) ([]submitter.Event, error) {
	args := t.Called(clock, nodeLister, metrics)
	return args.Get(0).([]submitter.Event), nil
}

