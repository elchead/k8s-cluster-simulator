from collections import defaultdict
import pathlib
import os
import json
import pandas
from pprint import pprint


def read_to_panda(strdata, config_name):
    d = json.loads(strdata)
    d["max_usage_gb"] = max(d["zone_max_usage_gb"].values())
    subd = {
        key: d[key]
        for key in d.keys()
        if key
        not in [
            "zone_mean_usage_gb",
            "zone_max_usage_gb",
            "provisions",
            "most_consuming_jobs",
            "migrated_pods",
            "total_job_time",
        ]
    }

    pd = pandas.DataFrame(subd, [str(config_name) + "%"])
    pd.columns = [
        "Job count",
        "Migration count",
        "Job time",
        "Migration time [s]",
        "Mean memory usage [GB]",
        "Mean memory usage [%]",
        "Provision count",
        "Max node usage [GB]",
    ]

    return pd


# def concat_pd():
#     p1 = read_to_panda(strdata, "First")
#     p2 = pandas.concat([p1, read_to_panda(data2, "Second")])
#     print(p2)  # .to_latex())


def get_subdirs(dir):
    p = pathlib.Path(dir)
    return [f for f in p.iterdir() if f.is_dir() and f.stem.startswith("t")]


def evaluate_migrated_pods(path, seed):
    paths = get_subdirs(path)

    fpath = f"{path}/{seed}/mig-report.txt"
    with open(fpath, "r") as f:
        strdata = f.read()
        d = json.loads(strdata)
        pprint(d["migrated_pods"])


def evaluate_tables(path, seed):
    pd = defaultdict(pandas.DataFrame)
    paths = get_subdirs(path)
    for path in paths:
        fpath = path.joinpath(f"{seed}/mig-report.txt")
        with open(fpath, "r") as f:
            strdata = f.read()
            config = os.path.basename(path)
            configls = config.split("-")
            req = configls[-1]
            mig = configls[2]
            threshold = int(configls[0][1:])
            try:
                res = read_to_panda(strdata, threshold)
                key = f"Requester:{req}; Migrator:{mig}"
                pd[key] = pandas.concat([pd[key], res])
            except Exception as e:
                print("SEED", seed, "failed", config, e)

    return pd


def evaluate_seed_tables(path, seedmax):
    pd = pandas.DataFrame()  # = defaultdict(pandas.DataFrame)
    # paths = get_subdirs(path)
    path = pathlib.Path(path)
    for seed in range(1, seedmax + 1):
        fpath = path.joinpath(f"{seed}/mig-report.txt")
        with open(fpath, "r") as f:
            strdata = f.read()
            config = os.path.basename(path)
            configls = config.split("-")
            req = configls[-1]
            mig = configls[2]
            threshold = int(configls[0][1:])
            try:
                res = read_to_panda(strdata, threshold)
            except:
                print("SEED", seed, "failed", threshold)
            key = f"Requester:{req}; Migrator:{mig}"
            pd = pandas.concat([pd, res])
    # print_table(pd)
    return pd


def evaluate_seed_tables_no_config(path, seedmax):
    pd = pandas.DataFrame()  # = defaultdict(pandas.DataFrame)
    # paths = get_subdirs(path)
    path = pathlib.Path(path)
    for seed in range(1, seedmax + 1):
        fpath = path.joinpath(f"{seed}/mig-report.txt")
        with open(fpath, "r") as f:
            strdata = f.read()
            try:
                res = read_to_panda(strdata, seed)
                pd = pandas.concat([pd, res])
            except Exception as e:
                print(e, "seed", seed, "failed", fpath)
    # print_table(pd)
    return pd


def evaluate_req_tables(path, seedmax):
    pd = pandas.DataFrame()  # = defaultdict(pandas.DataFrame)
    # paths = get_subdirs(path)
    path = pathlib.Path(path)
    for seed in range(1, seedmax + 1):
        fpath = path.joinpath(f"{seed}/mig-report.txt")
        with open(fpath, "r") as f:
            strdata = f.read()
            config = os.path.basename(path)
            fac = config[1:]  # config.split("-")
            try:
                res = read_to_panda(strdata, fac)
            except:
                print("SEED", seed, "failed", fac)
            pd = pandas.concat([pd, res])
    # print_table(pd)
    return pd


def print_table(pd):
    for k, v in pd.items():
        v.sort_index(inplace=True)
        print(k)
        print(v)
        print()
