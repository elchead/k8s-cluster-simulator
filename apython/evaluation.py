import json
from collections import defaultdict
import matplotlib.pyplot as plt

fname = "/Users/I545428/gh/controller-simulator/m-mig.log"


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


zones = ["zone2", "zone3", "zone4", "zone5"]


data = [json.loads(line) for line in open(fname, "r")]
for zone in zones:
    plt.figure()
    plt.title(zone)
    plt.xlabel("Time")
    plt.ylabel("Memory [Gb]")
    plt.legend()
    mems = get_pod_usage_on_node(zone, data)
    print()
    for pod, v in mems.items():
        plt.plot(v, label=pod)

plt.show()

