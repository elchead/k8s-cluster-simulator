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
	"time"

	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
)

type JobDeleter struct {
	jobs    []PodMemory
	endTime clock.Clock
	deleted map[string]bool
}

func NewJobDeleterWithEndtime(jobs []PodMemory, endTime time.Time) *JobDeleter {
	return &JobDeleter{jobs: jobs, endTime: clock.NewClock(endTime),deleted: make(map[string]bool)}
}

func (s *JobDeleter) Submit(
	currentTime clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	events := make([]submitter.Event, 0, len(s.jobs))
	for _, job := range s.jobs {
		isDeleted := s.deleted[job.Name]
		if !isDeleted && !job.IsMigrating() && clock.NewClock(job.EndAt).BeforeOrEqual(currentTime) {
			events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: job.Name})
			s.deleted[job.Name] = true	
		}
	}

	if s.endTime.BeforeOrEqual(currentTime) {
		for _, pod := range s.jobs {
			_,isDeleted := s.deleted[pod.Name]
			if !isDeleted {
				events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: pod.Name})
			}
		}
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}

	return events, nil
}
