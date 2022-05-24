import pytest
from parsing import *


# def test_get_pod_usage_on_nodes():
#     fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
#     data = [json.loads(line) for line in open(fname, "r")]
#     assert [0.033884765625, 0.033025390625, 0.03300390625, 0.03903125, 0.03903125] == get_pod_usage_on_nodes(data)[
#         "o10n-worker-l-4g2hn-b6lvf"
#     ]["zone5"].memory
#     assert [10, 80, 150, 220, 290] == get_pod_usage_on_nodes(data)["o10n-worker-l-4g2hn-b6lvf"]["zone5"].time
