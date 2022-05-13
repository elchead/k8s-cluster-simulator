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

func (it *Iterator) Value() interface{} {
	if it.current == len(it.jobs) {
		return nil
	}
	return it.jobs[it.current]
}

func (it *Iterator) Next() interface{} {
	if it.current == len(it.jobs) {
		return nil
	}
	it.current++
	return it.Value()
}

func NewIterator(jobs []PodMemory) *Iterator {
	return &Iterator{jobs, 0}
}

type JobSubmitter struct {
	jobs       []PodMemory
	currentIdx int
	iterator   *Iterator
}

func NewJobSubmitter(jobs []PodMemory) *JobSubmitter {
	return &JobSubmitter{jobs: jobs, currentIdx: 0, iterator: NewIterator(jobs)}
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
	nextJob, ok := s.iterator.Value().(PodMemory)
	if !ok {
		return events, nil

	}

	jobTime := clock.NewClock(nextJob.StartAt)
	for jobTime.BeforeOrEqual(currentTime) {
		pod := CreatePod(nextJob)
		events = append(events, &submitter.SubmitEvent{Pod: pod})

		nextJob, ok = s.iterator.Next().(PodMemory)
		if !ok {
			break
		}
		jobTime = clock.NewClock(nextJob.StartAt)
	}

	if s.iterator.RemainingValues() == 0 {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}

	return events, nil
}
