from evaluation import *
import matplotlib.pyplot as plt

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
data = [json.loads(line) for line in open(fname, "r")]
for zone in zones:
    plt.figure()
    plt.title(zone)
    plt.xlabel("Time")
    plt.ylabel("Memory [Gb]")
    plt.legend()
    mems = get_pod_usage_on_node(zone, data)
    new_mems, migration_idxs = find_migration_points_and_merge_pods(mems)
    print(migration_idxs)
    for pod, v in new_mems.items():
        print(pod)
        plt.plot(v, label=pod, markevery=migration_idxs[pod], marker="x")

plt.show()
