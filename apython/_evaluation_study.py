from read_reports import *
import numpy as np

path = "/Users/I545428/gh/controller-simulator/evaluation/pods_760"
path2 = "/Users/I545428/gh/controller-simulator/evaluation/pods_2715"

jobs = [760, 2715]
jobseeds = [12, 15]
# ! Job analysis
# pd = evaluate_tables(path, 8)
# pd = evaluate_tables(path2, 16)

# ! Unscheduler impact
def unscheduler_impact(job):
    print("Job", job)
    for thresh in [0, 10, 20, 30]:
        seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/unscheduler/t{thresh}-m-big-enough-r-threshold"
        pd = evaluate_seed_tables(seed_path, 20)
        count_failure = 0
        maxs = pd.loc[:, "Max node usage [Gb]"]
        for d in maxs:
            if d > 450:
                count_failure += 1
        print(f"Threshold {thresh}: {count_failure} failures out of {len(maxs)}")


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
    print(f"Nomig_dynamic: {count_failure} failures out of {len(maxs)}")


def combi_req_unsched_rate(job, nbrJobs):
    print("Job", job)
    # for thresh in range(1, nbrJobs + 1):
    seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/combi_req_unsched"
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
    print(f"Combi req sched: {count_failure} failures out of {len(maxs)}")


def reqfac_impact(job):
    print("Job", job)
    for thresh in [0, 0.1, 0.25, 0.5, 0.75]:  # , 0.005]:
        seed_path = f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/reqfac/p{thresh}"
        pd = evaluate_req_tables(seed_path, 20)
        count_failure = 0
        maxs = pd.loc[:, "Max node usage [Gb]"]
        for d in maxs:
            if d > 450:
                count_failure += 1
        jtime = min(pd.loc[:, "Job time"])
        print(f"Request factor {thresh}: {count_failure} failures out of {len(maxs)}\t minimal jobtime: {jtime}")


# nomig_dynamic_failure_rate(760, 500)
# reqfac_impact(760)
# reqfac_impact(2715)
# combi_req_unsched_rate(760, 20)

# controller config
job = jobs[0]
# run combi (unscheduler included in main?)
# pick one failed scenario from:
# combi_req_unsched_rate(job, 20)
seed = jobseeds[0]
print("---", job)
# pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_7", seed)
# print_table(pd)
evaluate_migrated_pods(
    f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_7/t15-m-slope-r-threshold", seed
)

# print("ONLY", job)
# pdOnly = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/onlycontroller", seed)
# print_table(pdOnly)

# pd["Max node usage [Gb]"] >


# job = 760
# seed = 12  # changed from 19 to 12
# print("---", job)
# pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/controller_7_7", seed)
# print_table(pd)
# # for name, value in pd.items():
# #     del value["Job count"]
# #     del value["Job time"]
# #     del value["Mean memory usage [Gb]"]
# #     del value["Mean memory usage [%]"]
# #     del value["Provision count"]
# #     print(name)
# #     print(value.to_latex())

# print("ONLY", job)
# pd = evaluate_tables(f"/Users/I545428/gh/controller-simulator/evaluation/pods_{job}/onlycontroller", seed)
# print_table(pd)
