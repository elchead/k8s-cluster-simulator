from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt

plot = True

fname = "/Users/I545428/gh/controller-simulator/m-sim.log"
title = "with migration"
evaluate_sim(title, plot, fname)
plt.show()
# with open("/Users/I545428/gh/controller-simulator/evaluation/t10-m-max-r-slope/06-20T10-45-38/mig-sim.log", "r") as f:
#     evaluate_provisions(f)
