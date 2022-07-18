import pathlib
from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt
import sys
from pathlib import Path

args = sys.argv
fname = "/Users/I545428/gh/controller-simulator/m-sim.log"
plot = False  # True


def get_subdirs(dir):
    p = pathlib.Path(dir)
    return [f for f in p.iterdir() if f.is_dir()]


if len(args) > 1:
    fname = args[1]
    plot = True

title = "with migration"
paths = get_subdirs("/Users/I545428/gh/controller-simulator/evaluation/pods_760/controller_7_7")
# npaths = []
# for p in paths:
#     for fail in ["/41", "/74", "/96", "/246", "/362", "/418", "/459"]:
#         if str(p).endswith(fail):
#             npaths.append(p)
# paths = npaths
plot = True
# [Path("/Users/I545428/gh/controller-simulator/evaluation/pods_760/controller_t15_mslope/193")]  #
# print(paths)
original_stdout = sys.stdout
for i, fname in enumerate(paths):
    with open(fname.joinpath("12/mig-report.txt"), "w") as f:
        sys.stdout = f
        evaluate_sim(title, plot, fname.joinpath("12/mig-sim-json.log"), simlog=fname.joinpath("12/mig-sim.log"))
    sys.stdout = original_stdout
    print(i, fname.stem)
# plt.show()
# with open("/Users/I545428/gh/controller-simulator/evaluation/t10-m-max-r-slope/06-20T10-45-38/mig-sim.log", "r") as f:
#     evaluate_provisions(f)
