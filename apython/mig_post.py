import pathlib
from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt
import sys

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
paths = get_subdirs("/Users/I545428/gh/controller-simulator/evaluation/pods_760/controller_t15_mslope")
# print(paths)
original_stdout = sys.stdout
for i, fname in enumerate(paths):
    with open(fname.joinpath("mig-report.txt"), "w") as f:
        sys.stdout = f
        evaluate_sim(title, plot, fname.joinpath("mig-sim-json.log"), simlog=fname.joinpath("mig-sim.log"))
    sys.stdout = original_stdout
    print(i, fname.stem)
# plt.show()
# with open("/Users/I545428/gh/controller-simulator/evaluation/t10-m-max-r-slope/06-20T10-45-38/mig-sim.log", "r") as f:
#     evaluate_provisions(f)
