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
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/containerd/containerd/log"
	kubesim "github.com/elchead/k8s-cluster-simulator/pkg"
	"github.com/elchead/k8s-cluster-simulator/pkg/clock"
	"github.com/elchead/k8s-cluster-simulator/pkg/config"
	"github.com/elchead/k8s-cluster-simulator/pkg/jobparser"
	"github.com/elchead/k8s-cluster-simulator/pkg/migration"
	"github.com/elchead/k8s-cluster-simulator/pkg/queue"
	"github.com/elchead/k8s-cluster-simulator/pkg/scheduler"
	"github.com/elchead/k8s-migration-controller/pkg/monitoring"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
)
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

// configPath is the path of the config file, defaulting to "config".
var podDataFile string //=  // "./pods_760.json"
var simDurationMap = map[string]time.Duration{"./pods_760.json":5 * time.Hour + 10*time.Minute,"./pods_2715.json":10 * time.Hour + 10*time.Minute}

var configPath string
var checkerType string
var migPolicy string
var requestPolicy string
var useMigrator bool
var nodeFreeThreshold float64
var requestFactor float64

func init() {
	rootCmd.PersistentFlags().StringVar(&podDataFile, "file", "./pods_760.json", "path to pod data")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config", "config file (excluding file extension)")
	rootCmd.PersistentFlags().StringVar(&checkerType, "checker", "blocking", "blocking or concurrent")
	rootCmd.PersistentFlags().StringVar(&migPolicy, "migPolicy", "optimal", "Migration choice policy: optimal,max,big-enough")
	rootCmd.PersistentFlags().StringVar(&requestPolicy, "reqPolicy", "threshold", "Policy to request memory freeing: threshold,slope")
	rootCmd.PersistentFlags().BoolVar(&useMigrator, "useMigrator", false, "use migrator (default false)")
	rootCmd.PersistentFlags().Float64Var(&nodeFreeThreshold, "threshold", 45., "node free threshold in % (default 45.)")
	rootCmd.PersistentFlags().Float64Var(&requestFactor, "requestFactor", 0., "fraction of job sizing request as decimal (default 0.)")
}

var rootCmd = &cobra.Command{
	Use:   "k8s-cluster-simulator",
	Short: "k8s-cluster-simulator provides a virtual kubernetes cluster interface for evaluating your scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		if useMigrator {
			requestFactor = 0. // much more usage when no resource set at all! //.1 //.25 too big?
		} else {
			requestFactor = 1. 
		}

		ctx := newInterruptableContext()
		// 1. Create a KubeSim with a pod queue and a scheduler.
		queue := queue.NewPriorityQueue()
		sched := buildScheduler() // see below
		metricClient := migration.NewClient()

		conf, err := kubesim.ReadConfig(configPath)
		if useMigrator {
			for i, logger := range conf.MetricsLogger {
				conf.MetricsLogger[i].Dest = "m-" + logger.Dest
			}
		}
		if err != nil {
			log.L.Fatal("Failed to read config:", err)
		}
		simDuration,ok := simDurationMap[podDataFile]
		if !ok {
			log.L.Fatal("Failed to find sim duration for file:", podDataFile)
		}
		sim,_ := kubesim.NewKubeSim(conf, queue, sched,metricClient)
		startTime, _ := time.Parse(time.RFC3339, conf.StartClock)
		endTime := startTime.Add(simDuration)

		
		file, err := os.Open(podDataFile)
		if err != nil {
			log.L.Fatal("Failed to read pod file:", err)
		}
		jobs,err := jobparser.ParsePodMemoriesFromJson(file)
		assertSimStartBeforeAllJobStarts(jobs,sim.Clock.ToMetaV1().Time)
		// job := jobparser.FindJob("o10n-worker-l-2xs2w-c7hh4",jobs)
		// fmt.Println("MEM",job.Records[0].Usage)
		// jobs = []jobparser.PodMemory{*job}
		if err != nil {
			log.L.Fatal("Failed to parse jobs", err)
		}
		podFactory := jobparser.NewPodFactory(requestFactor) //jobparser.PodFactory{SetResources: !useMigrator}
		submitter := jobparser.NewJobSubmitterWithFactory(jobs,podFactory)
		sim.AddSubmitter("JobSubmitter", submitter)
		if useMigrator {
			log.L.Info("Setting migration threshold:",nodeFreeThreshold, "% ",  nodeFreeThreshold/100.)
			cluster := monitoring.NewClusterWithSize(getNodeSize(conf))
			requestPolicy := monitoring.NewRequestPolicy(requestPolicy, cluster, metricClient,nodeFreeThreshold)
			migrationPolicy := monitoring.NewMigrationPolicy(migPolicy,cluster,metricClient)
			migController := monitoring.NewController(requestPolicy, migrationPolicy)
			checker := monitoring.NewMigrationChecker(checkerType)
			sim.AddSubmitter("JobMigrator", migration.NewSubmitterWithJobsWithEndTimeFactory(migController,jobs,endTime,podFactory,checker))


			unscheduler := &migration.Unscheduler{EndTime:clock.NewClock(endTime),ThresholdDecimal: 1. - nodeFreeThreshold/100.,ReschedulableDistanceDecimal:.15}
			sim.AddSubmitter("NodeUnscheduler", unscheduler)
		}
		sim.AddSubmitter("JobDeleter", jobparser.NewJobDeleterWithEndtime(jobs, endTime))
		// 3. Run the main loop of KubeSim.
		//    In each execution of the loop, KubeSim
		//      1) stores pods submitted from the registered submitters to its queue,
		//      2) invokes scheduler with pending pods and cluster state,
		//      3) emits cluster metrics to designated location(s) if enabled
		//      4) progresses the simulated clock
		if err := sim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
			log.L.Fatal(err)
		}
	},
}

func buildScheduler() scheduler.Scheduler {
	// 1. Create a generic scheduler that mimics a kube-scheduler.
	sched := scheduler.NewGenericScheduler( /* preemption enabled */ false)

	// 2. Register extender(s)
	// sched.AddExtender(
	// 	scheduler.Extender{
	// 		Name:             "MyExtender",
	// 		Filter:           filterExtender,
	// 		Prioritize:       prioritizeExtender,
	// 		Weight:           1,
	// 		NodeCacheCapable: true,
	// 	},
	// )

	// 2. Register plugin(s)
	// Predicate
	sched.AddPredicate(predicates.CheckNodeConditionPred,predicates.CheckNodeMemoryPressurePredicate) // not used 
	sched.AddPredicate(predicates.CheckNodeUnschedulablePred,predicates.CheckNodeUnschedulablePredicate)
	sched.AddPredicate("GeneralPredicates", predicates.GeneralPredicates)
	// sched.AddPrioritizer()
	// Prioritizer
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "BalancedResourceAllocation",
		Map:    priorities.BalancedResourceAllocationMap,
		Reduce: nil,
		Weight: 1,
	})
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "LeastRequested",
		Map:    priorities.LeastRequestedPriorityMap,
		Reduce: nil,
		Weight: 1,
	})

	return &sched
}

func getNodeSize(conf *config.Config) float64 {
	nodeSzStr := conf.Cluster[0].Status.Allocatable["memory"]
	nodeSz, err := strconv.ParseFloat(nodeSzStr[:len(nodeSzStr)-1], 64)
	if err != nil {
		log.L.Fatal("Failed to parse node size:", err)
	}
	return nodeSz
}

func newInterruptableContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// SIGINT (Ctrl-C) and SIGTERM cancel kubesim.Run().
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	return ctx
}

// for test
func lifo(pod0, pod1 *v1.Pod) bool { // nolint
	return pod1.CreationTimestamp.Before(&pod0.CreationTimestamp)
}

func assertSimStartBeforeAllJobStarts(jobs []jobparser.PodMemory,startTime time.Time) {
	for _,job := range jobs {
		if job.StartAt.Before(startTime) {
			log.L.Fatalf("job %s started at %s before sim start at %s", job.Name, job.StartAt.String(), startTime.String())
		}
	}
}
