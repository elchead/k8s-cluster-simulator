from evaluation import *
from unittest.mock import patch, mock_open
import unittest.mock as mock


def open_mockfile(data):
    with patch("builtins.open", mock_open(read_data=data)) as mock_file:
        return open(mock_file)


data = """
2022/07/18 11:15:08 Skipping cmd  default/o10n-worker-l-ddfpp-5vqb6  with usage  57  to node  zone3  because  16.88888888888889  would exceed threshold
2022/07/18 11:15:08 Provision more nodes:  Migration of pods: [{default/o10n-worker-l-ddfpp-5vqb6 57 }] failed because nodes are full. no place to fullfill request {zone2 0}
time="2022-07-18T11:15:08+02:00" level=debug msg="push migration to queue:o10n-worker-l-bfz7d-7jj7p size 244 to node zone3 finishing at 2022-06-09 22:40:48 +0200 CEST"
time="2022-07-18T11:15:08+02:00" level=info msg="MigrationTime o10n-worker-l-bfz7d-7jj7p 818 starting 2022-06-09 22:27:10 +0200 CEST"
"""


def test_provision_requests():
    f = open_mockfile(data)
    assert [
        "2022/07/18 11:15:08 Provision more nodes:  Migration of pods: [{default/o10n-worker-l-ddfpp-5vqb6 57 }] failed because nodes are full. no place to fullfill request {zone2 0}\n"
    ] == get_provision_requests(f)
