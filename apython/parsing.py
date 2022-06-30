import json
from collections import defaultdict
from dataclasses import dataclass
from re import S
from job import bytesto
from typing import List
import copy
from dateutil.parser import parse


def get_memory(t, node):
    try:
        z2 = t["Nodes"][node]["TotalResourceUsage"]["memory"]
        # print(z2)
        return z2
    except:
        return str(0)


def get_zone_memory(data, name):
    z_mem = []
    for t in data:
        z2 = get_memory(t, name)
        z_mem.append(bytesto(z2))
    return z_mem


def get_zone_memory(data, name):
    z_mem = []
    for t in data:
        z2 = get_memory(t, name)
        z_mem.append(bytesto(z2))
    return z_mem


def get_node_time(data):
    res = []
    for t in data:
        res.append(parse(t["Clock"]))
    return res


def get_node(name, stamp):
    return stamp["Nodes"][name]


def get_running_pods(node):
    return node["RunningPodsNum"]


def get_node_usage(node):
    return node["TotalResourceUsage"]


# def get_pod_names(stamp):
#     return [name for name in get_pods(stamp)]


# def get_pod_usage_on_node(node, data):
#     pods = defaultdict(list)
#     for d in data:
#         for k, v in get_pods(d).items():
#             name = k.split("/")[1]
#             if v["Node"] == node:
#                 if "memory" not in v["ResourceUsage"]:
#                     # last entry is empty (0)
#                     pods[name].append(0)
#                     continue
#                 pods[name].append(bytesto(v["ResourceUsage"]["memory"]))
#     return pods


# def find_migration_points_and_merge_pods(pod_memories):
#     pod_migration_idxs = defaultdict(list)
#     # check and count prepended m's
#     new_pod_memories = defaultdict(list)
#     for pod, v in pod_memories.items():
#         if pod.startswith("m"):
#             count = 0
#             for i, l in enumerate(pod):
#                 if l == "m":
#                     count += 1
#                 else:
#                     break
#             originalpod = pod[count:]
#             if originalpod in pod_memories:
#                 pod_migration_idxs[originalpod].append(len(pod_memories[originalpod]) - 1)
#             else:
#                 print(pod, "original", originalpod, "not found")
#             # print("extend", pod_memories[originalpod])
#             new_pod_memories[originalpod].extend(v)

#     # new_pod_memories = {k: v for k, v in pod_memories.items() if not k.startswith("m")}
#     return new_pod_memories, pod_migration_idxs

