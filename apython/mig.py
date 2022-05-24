from parsing import *
from job import *
from evaluation import evaluate_jobs, load_data, plot_node_usage

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/m-sim.log"
print("Evaluate with migration")
data, jobs = load_data(fname)

evaluate_jobs(zones, data, jobs)
print("----")
plot_node_usage("with migration", data, zones)

