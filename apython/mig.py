from parsing import *
from job import *
from evaluation import evaluate_jobs, load_data

zones = ["zone2", "zone3", "zone4", "zone5"]

fname = "/Users/I545428/gh/controller-simulator/m-mig.log"
data, jobs = load_data(fname)

evaluate_jobs(zones, data, jobs)
