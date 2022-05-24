import matplotlib.pyplot as plt
from parsing import *
from job import *


def load_data(fname):
    data = [json.loads(line) for line in open(fname, "r")]
    rawjobs = get_pod_usage_on_nodes_dict(data)
    jobs = create_jobs_from_dict(rawjobs)
    return data, jobs


def plot_node_usage(title, data, zones):
    plt.figure()
    plt.title(title)
    plt.xlabel("Time")
    plt.ylabel("Memory [Gb]")
    for zone in zones:
        plt.plot(get_zone_memory(data, zone), label=zone)
    plt.legend()
    plt.show()


def init_plot_dict(zones):
    plots = {}
    axis = {}
    for z in zones:
        plots[z] = plt.figure()
        plots[z].suptitle(z)
        axis[z] = plots[z].add_subplot(1, 1, 1)
    return axis


def evaluate_jobs(zones, data, jobs: "dict[str,Job]", plot=False):
    axis = {}
    if plot:
        axis = init_plot_dict(zones)

    total_job_time = 0
    zone_mem_utilization = {}
    zone_max_utilization = {}
    total_nbr_mig = 0
    for jobname, job in jobs.items():
        total_nbr_mig += job.nbr_migrations
        jtime = job.get_execution_time()
        # print(jobname, "time:", jtime)
        total_job_time += jtime
        for zone, poddata in job.get_pod_runs_for_plot():
            zone_mem = get_zone_memory(data, zone)
            zone_mem_utilization[zone] = np.mean(zone_mem)
            zone_max_utilization[zone] = np.max(zone_mem)
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

    print("Total jobs:", len(jobs.values()))
    print("Total #migrations:", total_nbr_mig)
    print("Total job time [s]:", total_job_time)
    print("Zone mean usage [Gb]:", zone_mem_utilization)
    print("Zone max usage [Gb]:", zone_max_utilization)
