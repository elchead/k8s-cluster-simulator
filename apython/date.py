from collections import defaultdict
import datetime
from dateutil.parser import parse


def get_dates(data) -> defaultdict(list):
    dates = defaultdict(list)
    # first date
    # for pod, poddata in data[0]["Pods"].items():
    #     dates[pod].append(parse(poddata["BoundAt"]))
    for stamp in data:  # enumerate(data[1:]):
        for pod, poddata in stamp["Pods"].items():
            next_date = parse(poddata["BoundAt"]) + datetime.timedelta(seconds=poddata["ExecutedSeconds"])
            dates[pod].append(next_date)
    return dates


# def get_next_date(prior_clock, diff_executed_seconds) -> datetime.datetime:
#     return prior_clock + datetime.timedelta(seconds=diff_executed_seconds)


# def get_first_date(stamp, exec):
#     t = parse(stamp)
#     return t - datetime.timedelta(seconds=exec)
