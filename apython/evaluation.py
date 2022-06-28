import matplotlib.pyplot as plt
from parsing import *
from job import *
from plotting import *
import re

zones = ["zone2", "zone3", "zone4"]  # zone 5


def get_provision_requests(file) -> List[str]:
    res = []
    pattern = "fullfill"
    lines = file.readlines()
    for line in lines:
        match = pattern in line  # re.search(pattern, line)
        if match:
            str_res = line
            res.append(str_res)
    return res


def evaluate_sim(title, plot, fname, nbr_jobs=50):
    print(f"Evaluate {title}")
    data, jobs = load_data(fname)
    evaluate_jobs(zones, data, jobs, title, plot=plot, nbr_jobs=nbr_jobs)
    # print("o10n-worker-l-bfz7d-7jj7p")
    # print(jobs["o10n-worker-l-bfz7d-7jj7p"].node_data[0].memory)
    # print(jobs["o10n-worker-l-bfz7d-7jj7p"].node_data[1].memory)
    # print("other")
    # print(jobs["o10n-worker-l-7qvtp-r7nmj"].node_data[0].memory)
    # print(jobs["o10n-worker-l-7qvtp-r7nmj"].node_data[1].memory)
    # print("----")
    if plot:
        plot_node_usage_with_mig_markers(title, data, zones)
        # plot_node_usage(title, data, zones)
    try:
        with open("mig-sim.log") as f:
            evaluate_provisions(f)
    except Exception as e:
        print("Could not evaluate provisions", e)


def evaluate_provisions(f):
    provs = get_provision_requests(f)
    print("Provision requests:", len(provs))
    for p in provs:
        print(p, end="")


def load_data(fname):
    data = [json.loads(line) for line in open(fname, "r")]
    rawjobs = get_pod_usage_on_nodes_dict(data)
    jobs = create_jobs_from_dict(rawjobs)
    return data, jobs


def maximum(a, b):

    if a >= b:
        return a
    else:
        return b


class SimEvaluation:
    total_job_time = 0

    total_migration_time = 0
    migration_times = {}
    total_nbr_mig = 0

    zone_mem_utilization = {}
    zone_max_utilization = {}

    job_max_mem = defaultdict(float)

    def add_job_time(self, t):
        self.total_job_time += t

    def add_migration_time(self, t):
        self.total_migration_time += t

    def add_migration(self, jobname, nbr, time):
        if nbr > 0:
            self.migration_times[jobname] = nbr
            self.total_nbr_mig += nbr
            self.total_migration_time += time

    def total_migrations(self):
        return self.total_nbr_mig

    def migration_time(self):
        return self.total_migration_time

    def job_time(self):
        return self.total_job_time - self.total_migration_time

    def set_job_max_mem(self, jobname, maxmem):
        self.job_max_mem[jobname] = maximum(np.max(maxmem), self.job_max_mem[jobname])

    def set_zone_stats(self, zone, mem):
        self.zone_mem_utilization[zone] = np.mean(mem)
        self.zone_max_utilization[zone] = np.max(mem)

    def get_top_pods(self, nbr):
        self.job_max_mem = dict(sorted(self.job_max_mem.items(), key=lambda item: item[1], reverse=True))
        return list(self.job_max_mem.keys())[:nbr]

    def get_top_pods_consumption(self, nbr) -> "List[tuple[str,int]]":
        return list(self.job_max_mem.items())[:nbr]


def evaluate_jobs(zones, data, jobs: "dict[str,Job]", title, plot=False, nbr_jobs=None):
    res = SimEvaluation()
    total_jobs = len(jobs.values())
    if not nbr_jobs:
        nbr_jobs = total_jobs

    axis = {}
    fig = None
    if plot:
        fig, axis = init_plot_dict(title, zones)

    for jobname, job in jobs.items():
        res.add_job_time(job.get_execution_time())
        res.add_migration(jobname, job.nbr_migrations, job.get_migration_time())
        try:
            for zone, poddata in job.get_pod_runs_for_plot():
                res.set_job_max_mem(jobname, np.max(poddata.memory))
        except Exception as e:
            print(f"Job {job.name} failed. Reason: {e}")
    for zone in zones:
        res.set_zone_stats(zone, get_zone_memory(data, zone))

    print("Total jobs:", total_jobs)
    print("Total #migrations:", res.total_migrations())
    print("Total job time [s]:", res.job_time())
    print("Total migration time [s]:", res.migration_time())
    memmean = np.mean(list(res.zone_mem_utilization.values()))
    print("Mean memory usage [Gb]:", memmean, f"({memmean/450*100}%)")
    print("Zone mean usage [Gb]:", res.zone_mem_utilization)
    print("Zone max usage [Gb]:", res.zone_max_utilization)

    print("Migrated pods:")
    top_pods = res.get_top_pods(nbr_jobs)
    for jobname, job in jobs.items():
        if job.nbr_migrations > 0:
            print(
                jobname,
                "size [Gb]:",
                res.job_max_mem[jobname],
                "#migrations:",
                job.nbr_migrations,
                f"from {job.get_node(0)}",
                f"to {job.get_node(1)};\t",
                "migration time [s]:",
                job.get_migration_time(),
            )
        for zone, poddata in job.get_pod_runs_for_plot():
            if plot and jobname in top_pods:
                axis[zone].plot(
                    poddata.time, poddata.memory, markevery=poddata.migration_idx, label=jobname, marker="x"
                )
                axis[zone].set_xticks([])
    print("Most consuming jobs:\n", res.get_top_pods_consumption(nbr_jobs))

    if plot:
        for z in zones:
            axis[z].legend()

        t = title.replace(" ", "_")
        plt.savefig(f"pod_mem_{t}")

