import matplotlib.pyplot as plt
from parsing import *


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
