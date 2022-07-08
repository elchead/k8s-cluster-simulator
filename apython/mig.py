from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt
import sys

args = sys.argv
fname = "/Users/I545428/gh/controller-simulator/m-sim.log"
plot = False  # True

if len(args) > 1:
    fname = args[1]
    plot = True

title = "with migration"
evaluate_sim(title, plot, fname)
# plt.show()
# with open("/Users/I545428/gh/controller-simulator/evaluation/t10-m-max-r-slope/06-20T10-45-38/mig-sim.log", "r") as f:
#     evaluate_provisions(f)
