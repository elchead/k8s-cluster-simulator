import matplotlib.pyplot as plt
from parsing import *
from job import *

zones = ["zone2", "zone3", "zone4", "zone5"]


def evaluate_sim(title, plot, fname, nbr_jobs=50):
    print(f"Evaluate {title}")
    data, jobs = load_data(fname)
    evaluate_jobs(zones, data, jobs, plot=plot, nbr_jobs=nbr_jobs)
    print("----")
    if plot:
        plot_node_usage(title, data, zones)


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


def maximum(a, b):

    if a >= b:
        return a
    else:
        return b


def evaluate_jobs(zones, data, jobs: "dict[str,Job]", plot=False, nbr_jobs=None):
    axis = {}
    if plot:
        axis = init_plot_dict(zones)

    total_job_time = 0
    total_migration_time = 0
    zone_mem_utilization = {}
    zone_max_utilization = {}
    job_max_mem = defaultdict(float)
    total_nbr_mig = 0
    for jobname, job in jobs.items():

        jtime = job.get_execution_time()
        # print(jobname, "time:", jtime)
        total_job_time += jtime
        for zone, poddata in job.get_pod_runs_for_plot():
            job_max_mem[jobname] = maximum(np.max(poddata.memory), job_max_mem[jobname])
    for zone in zones:
        zone_mem = get_zone_memory(data, zone)
        zone_mem_utilization[zone] = np.mean(zone_mem)
        zone_max_utilization[zone] = np.max(zone_mem)

    job_max_mem = dict(sorted(job_max_mem.items(), key=lambda item: item[1], reverse=True))
    if not nbr_jobs:
        nbr_jobs = len(job_max_mem)
    top_pods = list(job_max_mem.keys())[:nbr_jobs]
    for jobname, job in jobs.items():
        nbr = job.nbr_migrations
        migtime = job.get_migration_time()
        total_migration_time += migtime
        if nbr > 0:
            print(
                jobname, "size [Gb]:", job_max_mem[jobname], "#migrations:", nbr, "migration time [s]:", migtime,
            )
        total_nbr_mig += nbr

        for zone, poddata in job.get_pod_runs_for_plot():
            if plot and jobname in top_pods:
                axis[zone].plot(
                    poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x"
                )

    if plot:
        for z in zones:
            axis[z].legend()
        plt.xlabel("Time")
        plt.ylabel("Memory [Gb]")
        plt.show()

    total_job_time -= total_migration_time
    print("Total jobs:", len(jobs.values()))
    print("Total #migrations:", total_nbr_mig)
    print("Total job time [s]:", total_job_time)
    print("Total migration time [s]:", total_migration_time)
    print("Zone mean usage [Gb]:", zone_mem_utilization)
    print("Zone max usage [Gb]:", zone_max_utilization)
    print("Most consuming jobs:\n", list(job_max_mem.items())[:nbr_jobs])
