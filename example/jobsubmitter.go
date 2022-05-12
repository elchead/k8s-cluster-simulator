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

package main

import (
	"io"

	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/metrics"
	"github.com/elchead/k8s-cluster-simulator/pkg/submitter"
)

type jobSubmitter struct {
	jobs       []jobparser.PodMemory
	currentIdx int
}

func newJobSubmitter(podMemCsvFile io.Reader) *jobSubmitter {
	podmems := jobparser.ParsePodMemories(podMemCsvFile)
	jobparser.SortPodMemoriesByTime(podmems)
	return &jobSubmitter{jobs: podmems, currentIdx: 0}
}

func (s *jobSubmitter) Submit(
	currentTime clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {

	nextJob := s.jobs[s.currentIdx]

	events := make([]submitter.Event, 0, len(s.jobs))
	for currentTime.Before(clock.NewClock(nextJob.StartAt)) {
		pod := jobparser.CreatePod(nextJob)
		events = append(events, &submitter.SubmitEvent{Pod: pod})
		s.currentIdx++
		if s.currentIdx == len(s.jobs) {
			break
		}
		nextJob = s.jobs[s.currentIdx]
	}

	// if s.podIdx > 0 { // Test deleting previously submitted pod
	// 	podName := fmt.Sprintf("pod-%d", s.podIdx-1)
	// 	events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: podName})
	// }

	// for i := 0; i < submissionNum; i++ {
	// 	s.podIdx++
	// }

	if s.currentIdx == len(s.jobs) {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}

	return events, nil
}
