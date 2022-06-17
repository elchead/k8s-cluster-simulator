import matplotlib.pyplot as plt
from parsing import *
from job import *


def plot_node_usage_with_mig_markers(title, data, zones):
    fig = plt.figure()
    plt.title(title)
    plt.xlabel("Time")
    plt.ylabel("Memory [Gb]")
    rawjobs = get_pod_usage_on_nodes_dict(data)
    color_dict = {"zone2": "b", "zone3": "y", "zone4": "g", "zone5": "r"}
    for zone in zones:
        mem = get_zone_memory(data, zone)
        plt.plot(mem, label=zone, c=color_dict[zone])

    # ax = fig.gca()
    # for i, p in enumerate(ax.get_lines()):  # this is the loop to change Labels and colors
    #     if p.get_label() in zones[:i]:  # check for Name already exists
    #         idx = zones.index(p.get_label())  # find ist index
    #         p.set_c(ax.get_lines()[idx].get_c())  # set color
    #         p.set_label("_" + p.get_label())  # hide label in auto-legend
    zone_markers = defaultdict(list)
    for name, pod in rawjobs.items():
        if count_m(name) > 0:
            for zone, poddata in pod.items():
                poddata.is_migrated = True
                zone_markers[zone].append(poddata.t_idx)
                plt.plot(poddata.t_idx, poddata.get_migration_size(), label=zone, marker="x", c=color_dict[zone])
    handles, labels = plt.gca().get_legend_handles_labels()
    by_label = dict(zip(labels, handles))
    plt.legend(by_label.values(), by_label.keys())
    # plt.legend()
    plt.savefig(title.replace(" ", "_"))

    plt.figure()
    plt.title("Slope " + title)
    for zone in zones:
        mem = get_zone_memory(data, zone)
        slope = np.diff(mem)
        plt.plot(slope, label=zone, c=color_dict[zone])
    plt.legend(by_label.values(), by_label.keys())
    plt.savefig("slope_" + title.replace(" ", "_"))


def plot_node_usage(title, data, zones):
    plt.figure()
    plt.title(title)
    plt.xlabel("Time")
    plt.ylabel("Memory [Gb]")
    for zone in zones:
        plt.plot(get_zone_memory(data, zone), label=zone)
    plt.legend()
    plt.savefig(title.replace(" ", "_"))


def init_plot_dict(title, zones):
    plots = {}
    axis = {}
    fig, axs = plt.subplots(2, len(zones) - 2, sharex=True)
    fig.suptitle(f"Pod memories ({title})")

    fig.text(0.5, 0.04, "Time", ha="center")
    fig.text(0.04, 0.5, "Memory [Gb]", va="center", rotation="vertical")
    for i, z in enumerate(zones):
        axs[int(i / 2), int(i % 2)].set_title(z)
        axis[z] = axs[int(i / 2), int(i % 2)]
    return fig, axis
