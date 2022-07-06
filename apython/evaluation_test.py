import pytest
from parsing import *
from unittest.mock import patch, mock_open
from evaluation import *
from job import Migration_durations
from dateutil.parser import parse
from datetime import datetime


# NOT used
def test_get_mig_time():
    data = 'time="2022-07-06T07:58:20+02:00" level=info msg="MigrationTime mo10n-worker-l-jq2pn-wq4l6 272 starting 2022-06-09 22:57:10 +0200 CEST'
    f = open_mockfile(data)
    set_migration_times(f)
    assert Migration_durations["o10n-worker-l-jq2pn-wq4l6"] == [272]
    assert Migration_times["o10n-worker-l-jq2pn-wq4l6"] == [
        datetime.strptime("2022-06-09 22:57:10 +0200 CEST", "%Y-%m-%d %H:%M:%S %z %Z")
    ]


def open_mockfile(data):
    with patch("builtins.open", mock_open(read_data=data)) as mock_file:
        return open(mock_file)


def test_provision_requests():
    data = """
time="2022-06-20T10:45:42+02:00" level=debug msg="Clock 2022-05-12T09:10:10+02:00"
2022/06/20 10:45:42 migrator requesting: {zone4 342.7892561983471}
time="2022-06-20T10:45:42+02:00" level=debug msg="Submit failed: migrator failed: problem during migration request: migration of pod (264.000000) on node zone4 does not fullfill request (342.789256)"
"""
    f = open_mockfile(data)
    assert get_provision_requests(f) == [
        'time="2022-06-20T10:45:42+02:00" level=debug msg="Submit failed: migrator failed: problem during migration request: migration of pod (264.000000) on node zone4 does not fullfill request (342.789256)"\n'
    ]


# def test_get_pod_usage_on_nodes():
#     fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
#     data = [json.loads(line) for line in open(fname, "r")]
#     assert [0.033884765625, 0.033025390625, 0.03300390625, 0.03903125, 0.03903125] == get_pod_usage_on_nodes(data)[
#         "o10n-worker-l-4g2hn-b6lvf"
#     ]["zone5"].memory
#     assert [10, 80, 150, 220, 290] == get_pod_usage_on_nodes(data)["o10n-worker-l-4g2hn-b6lvf"]["zone5"].time
