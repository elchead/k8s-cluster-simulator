from parsing import *
from job import *
from evaluation import *
import matplotlib.pyplot as plt

plot = True

fname = "/Users/I545428/gh/controller-simulator/m-sim.log"
title = "with migration"
evaluate_sim(title, plot, fname)
# plt.show()