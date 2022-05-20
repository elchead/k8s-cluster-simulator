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

package jobparser_test

import (
	"testing"
	"time"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
	"github.com/stretchr/testify/assert"
)

func TestDeleteJobs(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}, {Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}}}}
	sut := jobparser.NewJobDeleterWithEndtime(jobs, now)
	events, _ := sut.Submit(simTime, nil, nil)
	assert.Len(t, events, len(jobs)+1)
}

func TestDeleteOnlyOnce(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)
	endTime := now.Add(1 *time.Hour)

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}},EndAt: now},}
	sut := jobparser.NewJobDeleterWithEndtime(jobs, endTime)
	events, _ := sut.Submit(simTime, nil, nil)
	assert.Len(t, events, 1)
	events, _ = sut.Submit(simTime.Add(1.*time.Second), nil, nil)
	assert.Empty(t, events)
}

func TestDeleteOnlyEndedJob(t *testing.T) {
	now := time.Now()
	simTime := clock.NewClock(now)
	endTime := now.Add(1 *time.Hour)

	jobs := []jobparser.PodMemory{{Name: "j1", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}},EndAt: now},{Name: "j2", StartAt: now, Records: []jobparser.Record{{Time: now, Usage: 100.}},EndAt:endTime}}
	sut := jobparser.NewJobDeleterWithEndtime(jobs, endTime)
	events, _ := sut.Submit(simTime, nil, nil)
	assert.Len(t, events, 1)
	events, _ = sut.Submit(clock.NewClock(endTime), nil, nil)
	assert.Len(t, events, 2)
}

func isDeleteEvent(event submitter.Event) (ok bool) {
	_, ok = event.(*submitter.DeleteEvent)
	return
}
