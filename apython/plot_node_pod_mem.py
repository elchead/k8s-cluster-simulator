import json
import matplotlib.pyplot as plt
from parsing import *

# with open("./mig.log", "r") as f:


# fname = "../nomig.log"
# data = [json.loads(line) for line in open(fname, "r")]
# plt.title("current job sizing model")
# plt.xlabel("Time")
# plt.ylabel("Memory [Gb]")
# # print(bytesto(202849602216, "g"), "\n", 202849602216 / d)
# plt.plot(get_zone_memory(data, "zone2"), label="zone2")
# plt.plot(get_zone_memory(data, "zone3"), label="zone3")
# plt.plot(get_zone_memory(data, "zone4"), label="zone4")
# plt.plot(get_zone_memory(data, "zone5"), label="zone5")
# plt.legend()

plt.figure()
plt.title("with migration")
plt.xlabel("Time")
plt.ylabel("Memory [Gb]")
fname = "../m-mig.log"
data = [json.loads(line) for line in open(fname, "r")]
t = get_node_time(data)
plt.plot(t, get_zone_memory(data, "zone2"), label="zone2")
plt.plot(t, get_zone_memory(data, "zone3"), label="zone3")
plt.plot(t, get_zone_memory(data, "zone4"), label="zone4")
plt.plot(t, get_zone_memory(data, "zone5"), label="zone5")
plt.legend()


# Pod sum
# plt.figure()
# fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
# data = [json.loads(line) for line in open(fname, "r")]
# rawjobs = get_pod_usage_on_nodes(data)
# jobs = merge_jobs(rawjobs)
# jobs = adjust_time_stamps(jobs)
# sum_mem = np.zeros(4 * 60)
# for jobname, job in jobs.items():
#     for zone, poddata in job.get_pod_runs():
#         if zone == "zone3":
#             for didx, t in enumerate(poddata.time):
#                 idx = int(t / 60)
#                 sum_mem[idx] += poddata.memory[didx]

# plt.plot(sum_mem)
# plt.show()
# print("Hi")

