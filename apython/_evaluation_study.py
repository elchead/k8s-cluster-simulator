from read_reports import *
import numpy as np
from matplotlib import pyplot as plt
import matplotlib

latex = True
if latex:
    matplotlib.use("pgf")
    matplotlib.rcParams.update(
        {"pgf.texsystem": "pdflatex", "font.family": "serif", "text.usetex": True, "pgf.rcfonts": False,}
    )

path = "/Users/I545428/gh/controller-simulator/evaluation/pods_760/controller_7_18"
path2 = "/Users/I545428/gh/controller-simulator/evaluation/pods_2715"

jobs = [760, 2715]
scenarios = {760: "Scenario 1", 2715: "Scenario 2"}
jobseeds = [12, 15]

# ! Unscheduler impact
def unscheduler_impact(job, seeds=20):
    print("Job", job)
    failrates = {}
    jtimes = []
    for thresh in [0, 10, 15, 20, 30]:
        seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/unscheduler/t{thresh}"  # -m-big-enough-r-threshold"
        pd = evaluate_seed_tables_no_config(seed_path, seeds)
        count_failure = 0
        maxs = pd.loc[:, "Max node usage [Gb]"]
        for d in maxs:
            if d > 450:
                count_failure += 1
        print(f"Threshold {thresh}: {count_failure} failures out of {len(maxs)}")
        failrates[thresh] = count_failure / len(maxs) * 100
        jtime = min(pd.loc[:, "Job time"]) / 3600.0
        print(pd.loc[:, "Job time"][199:])
        jtimes.append(jtime)
    jtimes = np.array(jtimes) / jtimes[0] * 100
    # print(jtimes)
    pf = pandas.DataFrame.from_dict(failrates, orient="index")
    pf.columns = ["Failure rate [%]"]
    pf.insert(1, "Relative total job time [%]", jtimes)
    print(pf)
    print(pf.to_latex())


def nomig_dynamic_failure_rate(job, nbrJobs):
    print("Job", job)
    # for thresh in range(1, nbrJobs + 1):
    seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/nomig_dynamic"
    pd = evaluate_seed_tables_no_config(seed_path, nbrJobs)
    count_failure = 0
    failed_seeds = []
    maxs = pd.loc[:, "Max node usage [Gb]"]
    for d in maxs:
        if d >= 450:
            count_failure += 1

    df_mask = pd.loc[:, "Max node usage [Gb]"] >= 450
    positions = np.flatnonzero(df_mask)
    filtered_df = pd.iloc[positions]
    print("failed seeds", filtered_df)
    print(f"Nomig_dynamic: {count_failure} failures out of {len(maxs)}; {count_failure/len(maxs)*100}%")


def combi_req_unsched_rate(job, nbrJobs):
    print("Job", job)
    # for thresh in range(1, nbrJobs + 1):
    seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/combi_req_unsched"  # _r0.25"
    pd = evaluate_seed_tables_no_config(seed_path, nbrJobs)
    jtime = min(pd.loc[:, "Job time"]) / 3600.0
    print("Combined job time", jtime)
    # pf.insert(1, "Relative total job time [%]", jtimes)
    count_failure = 0
    failed_seeds = []
    maxs = pd.loc[:, "Max node usage [Gb]"]
    jtimes = []
    for d in maxs:
        if d >= 450:
            count_failure += 1

    df_mask = pd.loc[:, "Max node usage [Gb]"] >= 450
    positions = np.flatnonzero(df_mask)
    filtered_df = pd.iloc[positions]
    print("failed seeds", filtered_df)
    print(f"Combi req sched: {count_failure} failures out of {len(maxs)}; {count_failure/len(maxs)*100}")


def controller_rate(job, nbrJobs):
    print("Job", job)
    # for thresh in range(1, nbrJobs + 1):
    seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_t15_mslope_r0.25"
    pd = evaluate_seed_tables_no_config(seed_path, nbrJobs)
    count_failure = 0
    failed_seeds = []
    maxs = pd.loc[:, "Max node usage [Gb]"]
    for d in maxs:
        if d >= 450:
            count_failure += 1

    df_mask = pd.loc[:, "Max node usage [Gb]"] >= 450
    positions = np.flatnonzero(df_mask)
    filtered_df = pd.iloc[positions]
    print("failed seeds", filtered_df, maxs)
    print(f"Controller rate: {count_failure} failures out of {len(maxs)}; {count_failure/len(maxs)*100}")


def reqfac_impact(job, seeds=20):
    print("Job", job)
    threshs = [0, 1.0]  # [0, 0.1, 0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
    failure_rates = []
    jobtimes = []
    for thresh in threshs:  # [0, 0.1, 0.25, 0.5, 0.75]:  # , 0.005]:
        seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/reqfac/p{thresh}"
        pd = evaluate_req_tables(seed_path, seeds)
        count_failure = 0
        maxs = pd.loc[:, "Max node usage [Gb]"]
        for d in maxs:
            if d > 450:
                count_failure += 1
        jtime = min(pd.loc[:, "Job time"]) / 3600.0
        rate = count_failure / len(maxs) * 100
        failure_rates.append(rate)
        jobtimes.append(jtime)

        print(
            f"Request factor {thresh}: {count_failure} failures out of {len(maxs)}\t minimal jobtime: {jtime}; {rate}"
        )
    jobtimes = np.array(jobtimes) / jobtimes[0] * 100
    print(jobtimes)
    return threshs, failure_rates, jobtimes


# controller config
job = jobs[1]
# reqfac_impact(job, seeds=500)
# unscheduler_impact(job, seeds=500)
# job = jobs[0]
# reqfac_impact(job, seeds=100)
# unscheduler_impact(job, seeds=100)
# # unscheduler_impact(jobs[1], seeds=200)

# ## failure rate PLOT
# plt.figure()
# threshs, failure_rates, jobtimes = reqfac_impact(job, seeds=200)
# threshs, failure_rates2, jobtimes2 = reqfac_impact(jobs[1], seeds=100)
# plt.ylabel("Failure rate [%]")
# plt.xlabel("Job request factor")
# plt.plot(threshs, failure_rates, label=scenarios[jobs[0]])
# plt.plot(threshs, failure_rates2, label=scenarios[jobs[1]])
# plt.legend()
# plt.savefig("failure_request_factorN")
# plt.savefig("failure_request_factorN.pgf")

# ## jobtime rate PLOT
# plt.figure()
# plt.xlabel("Job request factor")
# plt.ylabel("Relative total job time [%]")
# plt.plot(threshs, jobtimes, label=scenarios[jobs[0]])
# plt.plot(threshs, jobtimes2, label=scenarios[jobs[1]])
# plt.legend()
# plt.savefig("jobtime_request_factor")
# plt.savefig("jobtime_request_factor.pgf")


# run combi (unscheduler included in main?)
# pick one failed scenario from:
# nomig_dynamic_failure_rate(job, 500)
# combi_req_unsched_rate(job, 500)
# controller_rate(job, 500)
seed = jobseeds[0]
# print("---", job)
# pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_7", seed)
# print_table(pd)
# evaluate_migrated_pods(
#     f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_7/t15-m-slope-r-threshold", seed
# )

# print("ONLY", job)
# pdOnly = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/onlycontroller", seed)
# print_table(pdOnly)

# pd["Max node usage [Gb]"] >


job = jobs[1]
seed = jobseeds[1]  # changed from 19 to 12
# print("---", job)
pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_18", seed)
for name, value in pd.items():
    del value["Job count"]
    del value["Job time"]
    del value["Mean memory usage [Gb]"]
    del value["Mean memory usage [%]"]
    # del value["Provision count"]
    # print(name)
    ls = name.strip(" ").split(";")
    req = ls[0].split(":")[1]
    mig = ls[1].split(":")[1]
    value.sort_index(inplace=True)
    print(
        value.to_latex(
            caption=f"{scenarios[job]}: Varying thresholds with {req} requester and {mig} migrator",
            label=f"param_job_r_{req}_m_{mig}",
        )
    )
# print_table(pd)

# print("ONLY", job)
# pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/onlycontroller", seed)
# print_table(pd)
plt.show()
