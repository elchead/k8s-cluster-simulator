from evaluation import *
import matplotlib.pyplot as plt

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
data = [json.loads(line) for line in open(fname, "r")]
rawjobs = get_pod_usage_on_nodes(data)
jobs = merge_jobs(rawjobs)


plots = {}
axis = {}
for z in zones:
    plots[z] = plt.figure()
    plots[z].suptitle(z)
    axis[z] = plots[z].add_subplot(1, 1, 1)


for jobname, job in jobs.items():
    for zone, poddata in job.nodes.items():
        if zone in zones:
            axis[zone].plot(poddata.memory, markevery=poddata.migration_idx, label=jobname)

# for zone in zones:
#     plt.figure()
#     plt.title(zone)
#     plt.xlabel("Time")
#     plt.ylabel("Memory [Gb]")
#     plt.legend()
#     mems = get_pod_usage_on_node(zone, data)
#     new_mems, migration_idxs = find_migration_points_and_merge_pods(mems)
#     print(migration_idxs)
#     for pod, v in new_mems.items():
#         print(pod)
#         plt.plot(v, label=pod, markevery=migration_idxs[pod], marker="x")

for z in zones:
    axis[z].legend()
plt.xlabel("Time")
plt.ylabel("Memory [Gb]")
plt.show()
