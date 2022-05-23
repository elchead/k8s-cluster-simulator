import json
from collections import defaultdict
from dataclasses import dataclass
from re import S
from typing import List
import numpy as np

maxRestarts = 10


def PodDataWith(time, memory):
    p = PodData()
    p.memory = memory
    p.time = time
    return p


class PodData:
    def __init__(self):
        self.memory = []
        self.time = []
        self.migration_idx = []
        self.checkpointed = False
        self.restored = False

    memory: "List[float]"
    time: "List[float]"
    migration_idx: "List[int]"


class Job:
    nodes: "dict[str,PodData]"
    name: str
    nbr_migrations: int
    node_order: "List[str]"

    def __init__(self):
        self.nodes = defaultdict(PodData)
        self.nbr_migrations = 0
        self.node_order = [""] * maxRestarts
        self.node_data = [None] * maxRestarts

    def get_pod_runs(self):
        data = [e for e in self.node_data if e]
        zones = [e for e in self.node_order if e != ""]
        return zip(zones, data)


def bytesto(bytes):
    d = 1 << 20
    res = 0.0
    if bytes[-1] == "k":
        res = int(bytes[:-1]) * 1024
    elif bytes[-1] == "M":
        return float(bytes[:-1]) / 1e3
    else:
        res = float(bytes)
    return res / d / 1e3


def get_node(name, stamp):
    return stamp["Nodes"][name]


def get_running_pods(node):
    return node["RunningPodsNum"]


def get_node_usage(node):
    return node["TotalResourceUsage"]


def get_pods(stamp):
    return stamp["Pods"]


def get_pod_names(stamp):
    return [name for name in get_pods(stamp)]


def get_pod_usage_on_node(node, data):
    pods = defaultdict(list)
    for d in data:
        for k, v in get_pods(d).items():
            name = k.split("/")[1]
            if v["Node"] == node:
                if "memory" not in v["ResourceUsage"]:
                    # last entry is empty (0)
                    pods[name].append(0)
                    continue
                pods[name].append(bytesto(v["ResourceUsage"]["memory"]))
    return pods


# USE
def get_pod_usage_on_nodes(data):
    pods = defaultdict(lambda: defaultdict(PodData))  # [job][node]
    # for d in data:
    #     for k, v in get_pods(d).items():
    #         job = k.split("/")[1]
    #         node = v["Node"]
    #         pods[job] = {node: PodData()}

    for d in data:
        for k, v in get_pods(d).items():
            job = k.split("/")[1]
            node = v["Node"]
            if "memory" not in v["ResourceUsage"]:  # last entry is empty (0)
                # pods[job][node].memory.append(0)
                continue
            pods[job][node].memory.append(bytesto(v["ResourceUsage"]["memory"]))
            pods[job][node].time.append(v["ExecutedSeconds"])

    return pods


def get_all_jobs(datatimestamp):
    jobs = defaultdict(Job)
    for d in datatimestamp:
        for rawname, poddata in get_pods(d).items():
            name = rawname.split("/")[1]
            if "memory" not in poddata["ResourceUsage"]:
                # last entry is empty (0)
                jobs[name].append(0)
                continue
            node = poddata["Node"]
            jobs[name].nodes[node]
            # jobs[name].append(bytesto(v["ResourceUsage"]["memory"]))
            jobs[name].memory.append(bytesto(poddata["ResourceUsage"]["memory"]))
            # jobs[name].name = name
            # jobs[name].nodes[node].
    return jobs


def find_migration_points_and_merge_pods(pod_memories):
    pod_migration_idxs = defaultdict(list)
    # check and count prepended m's
    new_pod_memories = defaultdict(list)
    for pod, v in pod_memories.items():
        if pod.startswith("m"):
            count = 0
            for i, l in enumerate(pod):
                if l == "m":
                    count += 1
                else:
                    break
            originalpod = pod[count:]
            if originalpod in pod_memories:
                pod_migration_idxs[originalpod].append(len(pod_memories[originalpod]) - 1)
            else:
                print(pod, "original", originalpod, "not found")
            # print("extend", pod_memories[originalpod])
            new_pod_memories[originalpod].extend(v)

    # new_pod_memories = {k: v for k, v in pod_memories.items() if not k.startswith("m")}
    return new_pod_memories, pod_migration_idxs


def merge_jobs(jobs):
    # check and count prepended m's
    new_jobs = defaultdict(Job)  # [job][node]
    for jobname, nodes in jobs.items():
        count = 0
        if jobname.startswith("m"):
            count = count_m(jobname)

            # add migration index to pod with less m
            for node, nodedata in nodes.items():
                if nodedata.memory:
                    old_node = list(jobs[jobname[1:]].keys())[0]
                    new_jobs[jobname[count:]].nodes[old_node].restored = True
                new_jobs[jobname[count:]].node_order[count] = node

            jobname = jobname[count:]
            if count > new_jobs[jobname].nbr_migrations:
                new_jobs[jobname].nbr_migrations = count

        for node, nodedata in nodes.items():
            if not jobname.startswith("m"):
                new_jobs[jobname].node_order[0] = node
            new_jobs[jobname].nodes[node].memory = nodedata.memory  # check if restarted on same node
            new_jobs[jobname].nodes[node].time = nodedata.time

            new_jobs[jobname].node_data[count] = PodDataWith(
                nodedata.time, nodedata.memory
            )  # check if restarted on same node

    return new_jobs


def adjust_time_stamps(jobs: "dict[str,Job]"):
    for jobname, job in jobs.items():
        for idx, data in enumerate(job.node_data[1:]):
            if data:
                last_time = jobs[jobname].node_data[idx].time[-1]
                new_time = np.array(data.time) + last_time
                data.time = new_time
                if data.time[0] != 0:
                    data.migration_idx = [0]

    return jobs


def count_m(job):
    count = 0
    for i, l in enumerate(job):
        if l == "m":
            count += 1
        else:
            break
    return count

