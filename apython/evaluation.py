import matplotlib.pyplot as plt
from parsing import *
from job import *


def load_data(fname):
    data = [json.loads(line) for line in open(fname, "r")]
    rawjobs = get_pod_usage_on_nodes_dict(data)
    jobs = create_jobs_from_dict(rawjobs)
    return data, jobs


def init_plot_dict(zones):
    plots = {}
    axis = {}
    for z in zones:
        plots[z] = plt.figure()
        plots[z].suptitle(z)
        axis[z] = plots[z].add_subplot(1, 1, 1)
    return axis


def evaluate_jobs(zones, data, jobs, plot=False):
    axis = {}
    if plot:
        axis = init_plot_dict(zones)

    total_job_time = 0
    zone_mem_utilization = {}
    for jobname, job in jobs.items():
        jtime = job.get_execution_time()
        print(jobname, "time:", jtime)
        total_job_time += jtime
        for zone, poddata in job.get_pod_runs_for_plot():
            zone_mem = get_zone_memory(data, zone)
            zone_mem_utilization[zone] = np.mean(zone_mem)
            if plot:
                axis[zone].plot(
                    poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x"
                )
    if plot:
        for z in zones:
            axis[z].legend()
        plt.xlabel("Time")
        plt.ylabel("Memory [Gb]")
        plt.show()

    print("Total job time [s]:", total_job_time)
    print("Zone usage [Gb]:", zone_mem_utilization)
