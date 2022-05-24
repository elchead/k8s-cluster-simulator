from parsing import *
import matplotlib.pyplot as plt
from job import *

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/mig.log"
data = [json.loads(line) for line in open(fname, "r")]
rawjobs = get_pod_usage_on_nodes_dict(data)
jobs = create_jobs_from_dict(rawjobs)
# jobs = adjust_time_stamps(jobs)


plots = {}
axis = {}
for z in zones:
    plots[z] = plt.figure()
    plots[z].suptitle(z)
    axis[z] = plots[z].add_subplot(1, 1, 1)


total_job_time = 0
for jobname, job in jobs.items():
    jtime = job.get_execution_time()
    print(jobname, "time:", jtime)
    total_job_time += jtime
    for zone, poddata in job.get_pod_runs_for_plot():
        axis[zone].plot(poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x")

print("Total job time", total_job_time, "s")
# for idx, poddata in enumerate(job.node_data):
#     if poddata:
#         zone = job.node_order[idx]
#         axis[zone].plot(poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x")

# for zone, poddata in job.nodes.items():
#     if zone in zones:
#         if job.node_order[1:] != ["", "", ""]:
#             print("Migration order", jobname, job.node_order)
#         if poddata.restored:
#             poddata.migration_idx = [0]
#         # if job.nbr_migrations != 0:
#         #     print("NBR", jobname, job.nbr_migrations)
#         axis[zone].plot(poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x")

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
#         plt.plot(v, label=pod)

for z in zones:
    axis[z].legend()

plt.xlabel("Time")
plt.ylabel("Memory [Gb]")
plt.show()
