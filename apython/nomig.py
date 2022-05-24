from parsing import *
from job import *
from evaluation import evaluate_jobs, load_data, plot_node_usage

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/sim.log"
data, jobs = load_data(fname)
print("Evaluate status quo")
evaluate_jobs(zones, data, jobs)
print("----")
plot_node_usage("no migration", data, zones)

