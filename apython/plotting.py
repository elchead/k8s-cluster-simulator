import matplotlib.pyplot as plt
from parsing import *
from job import *

from matplotlib.cm import get_cmap
from matplotlib.pyplot import cm
import matplotlib

from datetime import datetime, timedelta
import matplotlib.dates as mdates

# if need more colors use: https://stackoverflow.com/a/25730396/10531075
name = "tab20"
cmap = get_cmap(name)  # type: matplotlib.colors.ListedColormap
colors = cmap.colors

dpi = 200

latex = True
if latex:
    matplotlib.use("pgf")
    matplotlib.rcParams.update(
        {"pgf.texsystem": "pdflatex", "font.family": "serif", "text.usetex": True, "pgf.rcfonts": False,}
    )


def plot_node_usage_with_mig_markers(title, data, zones):
    fig = plt.figure()
    # plt.title(title)
    plt.xlabel("Time")
    plt.ylabel("Memory [GB]")
    plt.ylim(0, 450)
    rawjobs, _ = get_pod_usage_on_nodes_dict(data)
    color_dict = {"zone2": "b", "zone3": "y", "zone4": "g", "zone5": "r"}
    t = get_node_time(data)
    # print("TIME", t)
    # plt.plot(t, range(len(t)))
    for zone in zones:
        mem = get_zone_memory(data, zone)
        plt.plot(t, mem, label=zone, c=color_dict[zone])

    # ax = fig.gca()
    # for i, p in enumerate(ax.get_lines()):  # this is the loop to change Labels and colors
    #     if p.get_label() in zones[:i]:  # check for Name already exists
    #         idx = zones.index(p.get_label())  # find ist index
    #         p.set_c(ax.get_lines()[idx].get_c())  # set color
    #         p.set_label("_" + p.get_label())  # hide label in auto-legend
    zone_markers = defaultdict(list)
    for name, pod in rawjobs.items():
        if count_m(name) > 0:
            prior_pod = "m" * (count_m(name) - 1) + name[count_m(name) :]
            prior_zone = list(rawjobs[prior_pod].keys())[0]
            for zone, poddata in pod.items():
                poddata.is_migrated = True
                plt.plot(
                    t[poddata.t_idx] - timedelta(seconds=get_migration_time(poddata.get_migration_size())),
                    poddata.get_migration_size(),
                    label=prior_zone,
                    marker="x",
                    c=color_dict[prior_zone],
                )
                zone_markers[prior_zone].append(poddata.t_idx)
    handles, labels = plt.gca().get_legend_handles_labels()
    by_label = dict(zip(labels, handles))
    plt.legend(by_label.values(), by_label.keys())
    # plt.legend()
    plt.savefig(title.replace(" ", "_") + ".pdf")
    # plt.savefig(title.replace(" ", "_") + ".pgf", dpi=dpi)
    # plt.figure()
    # plt.title("Slope " + title)
    # for zone in zones:
    #     mem = get_zone_memory(data, zone)
    #     slope = np.diff(mem)
    #     plt.plot(slope, label=zone, c=color_dict[zone])
    # plt.legend(by_label.values(), by_label.keys())
    # plt.savefig("slope_" + title.replace(" ", "_"), dpi=dpi)


def plot_node_usage(title, data, zones):
    plt.figure()
    plt.title(title)
    plt.xlabel("Time")
    plt.ylabel("Memory [GB]")
    for zone in zones:
        plt.plot(get_zone_memory(data, zone), label=zone)
    plt.legend()
    plt.savefig(title.replace(" ", "_"), dpi=dpi)
    if latex:
        plt.savefig(title.replace(" ", "_") + ".pgf", dpi=dpi)


def init_plot_dict(title, zones):
    plots = {}
    axisdict = {}

    fig = None
    axs = []
    if len(zones) > 3:
        fig, axs = plt.subplots(2, len(zones) - 2, sharex=True, sharey=True)
    else:
        fig, axs = plt.subplots(len(zones), 1, sharex=True, sharey=True)
    fig.set_figheight(6)
    fig.set_figwidth(7)
    # fig.suptitle(f"Pod memories ({title})")

    fig.text(0.5, 0.04, "Time [h]", ha="center")
    fig.text(0.04, 0.5, "Memory [GB]", va="center", rotation="vertical")
    for i, z in enumerate(zones):
        if len(zones) > 3:
            axs[int(i / 2), int(i % 2)].set_title(z)
            axisdict[z] = axs[int(i / 2), int(i % 2)]
        else:
            axs[i].set_title(z)
            axs[i].set_ylim([0, 450])
            axs[i].yaxis.set_ticks(range(0, 451, 150))
            # axs[i].xaxis.set_major_locator(mdates.DayLocator(interval=3))  # to get a tick every 15 minutes
            axs[i].xaxis.set_major_formatter(
                mdates.DateFormatter("%H")
                # mdates.ConciseDateFormatter(axs[i].xaxis.get_major_locator())
            )
            axisdict[z] = axs[i]
            axisdict[z].set_prop_cycle(color=colors)
    return fig, axisdict
