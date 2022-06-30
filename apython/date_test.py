import datetime
import json
import pytest
from dateutil.parser import parse
import matplotlib.pyplot as plt
from evaluation_test import open_mockfile

# from evaluation import load_data
from date import *

txt = """\
{"Clock":"2022-06-09T22:02:20+02:00","Nodes":{"zone2":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":178,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1360","memory":"74849417213"}},"zone3":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":153,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1192","memory":"146210884943"}},"zone4":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":174,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1344","memory":"179582740475"}}},"Pods":{"default/o10n-worker-l-29d6p-k2lkp":{"ResourceRequest":{},"ResourceLimit":{},"ResourceUsage":{"cpu":"8","memory":"25688064"},"BoundAt":"2022-06-09T22:00:00+02:00","Node":"zone4","ExecutedSeconds":80,"Priority":0,"Status":"Ok"}}}
{"Clock":"2022-06-09T22:02:21+02:00","Nodes":{"zone2":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":178,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1360","memory":"74849417213"}},"zone3":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":153,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1192","memory":"146210884943"}},"zone4":{"Allocatable":{"cpu":"120","memory":"450G","pods":"1k"},"RunningPodsNum":174,"TerminatingPodsNum":0,"FailedPodsNum":0,"TotalResourceRequest":{},"TotalResourceUsage":{"cpu":"1344","memory":"179582740475"}}},"Pods":{"default/o10n-worker-l-29d6p-k2lkp":{"ResourceRequest":{},"ResourceLimit":{},"ResourceUsage":{"cpu":"8","memory":"1"},"BoundAt":"2022-06-09T22:00:00+02:00","Node":"zone4","ExecutedSeconds":140,"Priority":0,"Status":"Ok"}}}\
"""


def test_plot_date():
    stamp = "2022-06-09T22:16:20+02:00"
    t = parse(stamp)
    dates = [t + datetime.timedelta(minutes=1 * i) for i in range(10)]
    plt.plot(dates, range(10))
    plt.show()


def test_read_times():
    f = open_mockfile(txt)
    data = [json.loads(line) for line in f]
    dates = get_dates(data)
    # TODO no BoundAt included (add memory 0?)
    assert parse("2022-06-09T22:01:20+02:00") == dates["default/o10n-worker-l-29d6p-k2lkp"][0]
    assert parse("2022-06-09T22:02:20+02:00") == dates["default/o10n-worker-l-29d6p-k2lkp"][1]

