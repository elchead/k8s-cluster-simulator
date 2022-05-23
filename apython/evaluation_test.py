import pytest
from evaluation import *


def test_find_migration_times():
    pod_mem = {"o10wo": [1, 2, 3], "mo10wo": [4, 5, 6], "mmo10wo": [7, 8, 9]}
    new_pod_mem, idxs = find_migration_points_and_merge_pods(pod_mem)
    assert idxs == {"o10wo": [2, 5]}
    assert new_pod_mem == {"o10wo": [1, 2, 3, 4, 5, 6, 7, 8, 9]}


def test_get_job():
    raw_pod_mem = {
        "o10wo": {"zone2": PodData([1, 2, 3], [10, 20, 30])},
        "mo10wo": {"zone3": PodData([4, 5, 6], [10, 20, 30])},
        "mmo10wo": {"zone4": PodData([7, 8, 9], [10, 20, 30])},
    }
    j = merge_jobs(raw_pod_mem)
    assert j["o10wo"]["zone2"].memory == [1, 2, 3]
    assert j["o10wo"]["zone2"].time == [10, 20, 30]
    assert j["o10wo"]["zone3"].memory == [4, 5, 6]
    assert j["o10wo"]["zone3"].time == [10, 20, 30]
    assert j["o10wo"]["zone4"].memory == [7, 8, 9]


def test_get_pod_usage_on_nodes():
    fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
    data = [json.loads(line) for line in open(fname, "r")]
    assert [0.033884765625, 0.033025390625, 0.03300390625, 0.03903125, 0.03903125] == get_pod_usage_on_nodes(data)[
        "o10n-worker-l-4g2hn-b6lvf"
    ]["zone5"].memory
    assert [10, 80, 150, 220, 290] == get_pod_usage_on_nodes(data)["o10n-worker-l-4g2hn-b6lvf"]["zone5"].time
