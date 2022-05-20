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

package jobparser

import (
	"io"

	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
)

type Iterator struct {
	jobs    []PodMemory
	current int
}

func (it *Iterator) RemainingValues() int {
	return len(it.jobs) - it.current
}

func (it *Iterator) Value() PodMemory {
	return it.jobs[it.current]
}


func (it *Iterator) Push(job PodMemory) {
	it.jobs = append(it.jobs, job)
}

func (it *Iterator) Next() bool {
	if it.current >= len(it.jobs)-1 {
		it.current = len(it.jobs)
		return false
	} else {
		it.current++
		return true
	}
}

func (it *Iterator) ExistNext() bool {
	if it.current == len(it.jobs) {
		return false
	} else {
		return true
	}
}

func NewIterator(jobs []PodMemory) *Iterator {
	return &Iterator{jobs, 0}
}

type JobSubmitter struct {
	jobs       []PodMemory
	currentIdx int
	iterator   *Iterator
	factory PodFactory
}

func NewJobSubmitter(jobs []PodMemory) *JobSubmitter {
	return &JobSubmitter{jobs: jobs, currentIdx: 0, iterator: NewIterator(jobs),factory: PodFactory{SetResources: true}}
}

func NewJobSubmitterWithFactory(jobs []PodMemory,podfactory PodFactory) *JobSubmitter {
	return &JobSubmitter{jobs: jobs, currentIdx: 0, iterator: NewIterator(jobs),factory: podfactory}
}

func NewJobSubmitterFromFile(podMemCsvFile io.Reader) *JobSubmitter {
	podmems := ParsePodMemories(podMemCsvFile)
	return NewJobSubmitter(podmems)
}

func (s *JobSubmitter) Submit(
	currentTime clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {

	events := make([]submitter.Event, 0, s.iterator.RemainingValues()+1)
	if !s.iterator.ExistNext() {
		events = append(events, &submitter.TerminateSubmitterEvent{})
		return events, nil
	}

	for s.iterator.ExistNext() {
		nextJob := s.iterator.Value()
		jobTime := clock.NewClock(nextJob.StartAt)
		if jobTime.BeforeOrEqual(currentTime) {
			pod := s.factory.New(nextJob)
			events = append(events, &submitter.SubmitEvent{Pod: pod})
			s.iterator.Next()
		} else {
			break
		}
	}

	if s.iterator.RemainingValues() == 0 {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}

	return events, nil
}
