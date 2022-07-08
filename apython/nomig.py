from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt
import sys

args = sys.argv
fname = "/Users/I545428/gh/controller-simulator/sim.log"
plot = False  # True

if len(args) > 1:
    fname = args[1]
    plot = True

plot = False  # True
title = "status quo"
evaluate_sim(title, plot, fname)
# plt.show()

