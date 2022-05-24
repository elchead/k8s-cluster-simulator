import json
from collections import defaultdict
from dataclasses import dataclass
from re import S
from typing import List
import numpy as np
import copy

maxRestarts = 10


class PodData:
    @classmethod
    def withdata(cls, time, memory):
        p = cls()
        p.memory = memory
        p.time = time
        return p

    def __init__(self):
        self.memory = []
        self.time = []
        self.migration_idx = []

    def get_execution_time(self):
        return self.time[-1]


class Job:
    nodes: "dict[str,PodData]"
    name: str
    nbr_migrations: int
    node_order: "List[str]"
    node_data: "List[PodData]"

    def __init__(self):
        self.nodes = defaultdict(PodData)
        self.nbr_migrations = 0
        self.node_order = [""] * maxRestarts
        self.node_data = [None] * maxRestarts  # PodData

    def add_pod(self, podname, node, nodedata: PodData):
        count = count_m(podname)
        self._count_migration(count)
        self.node_order[count] = node
        self.node_data[count] = PodData.withdata(nodedata.time, nodedata.memory)

    def _count_migration(self, count):
        if count > self.nbr_migrations:
            self.nbr_migrations = count

    def get_pod_runs_for_plot(self):
        data = [e for e in get_shifted_timestamps(self.node_data) if e]
        zones = [e for e in self.node_order if e != ""]
        return zip(zones, data)

    def get_execution_time(self):
        total = 0
        for poddata in self.node_data:
            if poddata:
                total += poddata.get_execution_time()
        return total


def get_memory(t, node):
    z2 = t["Nodes"][node]["TotalResourceUsage"]["memory"]
    try:
        z2 = int(z2)
    except:
        if z2[-1] == "k":
            z2 = int(z2[:-1]) * 8192
    return z2


def bytestoOld(bytes, bsize=1024):
    """convert bytes to megabytes, etc.
       sample code:
           print('mb= ' + str(bytesto(314575262000000, 'm')))
       sample output: 
           mb= 300002347.946
    """

    # a = {"k": 1, "m": 2, "g": 3, "t": 4, "p": 5, "e": 6}
    # r = float(bytes)
    # for i in range(a[to]):
    #     r = r / bsize
    d = 1 << 20

    return bytes / d / 1e3


def get_zone_memory(data, name):
    z_mem = []
    for t in data:
        z2 = get_memory(t, name)
        z_mem.append(bytestoOld(z2))
    return z_mem


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


def merge_jobs(job_node_dict):
    new_jobs = defaultdict(Job)  # [job][node]
    for podname, nodes in job_node_dict.items():
        count = count_m(podname)
        jobname = podname[count:]
        for node, nodedata in nodes.items():
            new_jobs[jobname].add_pod(podname, node, nodedata)
    return new_jobs


def get_shifted_timestamps(p: "List[PodData]"):
    cp = copy.deepcopy(p)
    for prior_idx, data in enumerate(cp[1:]):
        if data:
            last_time = cp[prior_idx].time[-1]
            new_time = np.array(data.time) + last_time
            data.time = new_time
            if data.time[0] != 0:
                data.migration_idx = [0]
    return cp


def count_m(job):
    count = 0
    for i, l in enumerate(job):
        if l == "m":
            count += 1
        else:
            break
    return count
