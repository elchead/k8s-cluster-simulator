from collections import defaultdict
from typing import List
import copy
import numpy as np
import math

maxRestarts = 10


def get_migration_time(gbSz: float):
    return math.ceil(3.3506 * gbSz)


def create_jobs_from_dict(job_node_dict):
    new_jobs = defaultdict(Job)
    for podname, nodes in job_node_dict.items():
        count = count_m(podname)
        jobname = podname[count:]
        for node, nodedata in nodes.items():
            new_jobs[jobname].add_pod(podname, node, nodedata)
    return new_jobs


def get_pod_usage_on_nodes_dict(data):
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


class PodData:
    @classmethod
    def withdata(cls, time, memory, is_migrated=False):
        p = cls()
        p.memory = memory
        p.time = time
        p.is_migrated = is_migrated
        return p

    def __init__(self):
        self.memory = []
        self.time = []
        self.migration_idx = []
        self.is_migrated = False

    def get_execution_time(self):
        return self.time[-1]

    def get_migration_time(self):
        if not self.is_migrated:
            return 0
        else:
            return get_migration_time(self.memory[0])


class Job:
    nodes: "dict[str,PodData]"
    name: str
    nbr_migrations: int
    node_order: "List[str]"
    node_data: "List[PodData]"

    def __init__(self, name=""):
        self.nodes = defaultdict(PodData)
        self.nbr_migrations = 0
        self.node_order = [""] * maxRestarts
        self.node_data = [None] * maxRestarts  # PodData
        self.name = self._set_name(name)

    def add_pod(self, podname, node, nodedata: PodData):
        count = count_m(podname)
        self._count_migration(count)
        self._set_name(podname)
        self.node_order[count] = node
        is_migrated = count > 0
        self.node_data[count] = PodData.withdata(nodedata.time, nodedata.memory, is_migrated=is_migrated)

    def _set_name(self, name):
        count = count_m(name)
        self.name = name[count:]

    def _count_migration(self, count):
        if count > self.nbr_migrations:
            self.nbr_migrations = count

    def get_pod_runs_for_plot(self):
        data = [e for e in add_migration_idx(get_shifted_timestamps(self.node_data)) if e]
        zones = [e for e in self.node_order if e != ""]
        return zip(zones, data)

    def get_execution_time(self):
        total = 0
        for poddata in self.node_data:
            if poddata:
                total += poddata.get_execution_time()
        return total

    def get_migration_time(self):
        total = 0
        for poddata in self.node_data:
            if poddata:
                total += poddata.get_migration_time()
        return total


def get_shifted_timestamps(p: "List[PodData]"):
    cp = copy.deepcopy(p)
    for prior_idx, data in enumerate(cp[1:]):
        if data:
            last_time = cp[prior_idx].time[-1]
            new_time = np.array(data.time) + last_time
            data.time = new_time
    return cp


def add_migration_idx(p):
    last_restarted_idx = -1
    for prior_idx, data in enumerate(p[1:]):
        if data:
            # was restarted
            if data.time[0] != 0:
                data.migration_idx = [0]
                last_restarted_idx = prior_idx + 1
    for idx, data in enumerate(p):
        if last_restarted_idx > idx:
            data.migration_idx.append(len(data.memory) - 1)
    return p


def get_pods(stamp):
    return stamp["Pods"]


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


def count_m(job):
    count = 0
    for i, l in enumerate(job):
        if l == "m":
            count += 1
        else:
            break
    return count