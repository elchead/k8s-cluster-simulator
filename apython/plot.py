import json
import matplotlib.pyplot as plt

# with open("./mig.log", "r") as f:


def get_memory(t, node):
    z2 = t["Nodes"][node]["TotalResourceUsage"]["memory"]
    try:
        z2 = int(z2)
    except:
        if z2[-1] == "k":
            z2 = int(z2[:-1]) * 8192
    return z2


def bytesto(bytes, to, bsize=1024):
    """convert bytes to megabytes, etc.
       sample code:
           print('mb= ' + str(bytesto(314575262000000, 'm')))
       sample output: 
           mb= 300002347.946
    """

    # a = {"k": 1, "m": 2, "g": 3, "t": 4, "p": 5, "e": 6}
    # r = float(bytes)
    # for i in range(a[to]):
    #     r = r / bsize
    d = 1 << 20

    return bytes / d / 1e3


def get_zone_memory(data, name):
    z_mem = []
    for t in data:
        z2 = get_memory(t, name)
        z_mem.append(bytesto(z2, "g"))
    return z_mem


fname = "../nomig.log"
data = [json.loads(line) for line in open(fname, "r")]
plt.title("current job sizing model")
plt.xlabel("Time")
plt.ylabel("Memory [Gb]")
# print(bytesto(202849602216, "g"), "\n", 202849602216 / d)
plt.plot(get_zone_memory(data, "zone2"), label="zone2")
plt.plot(get_zone_memory(data, "zone3"), label="zone3")
plt.plot(get_zone_memory(data, "zone4"), label="zone4")
plt.plot(get_zone_memory(data, "zone5"), label="zone5")
plt.legend()

plt.figure()
plt.title("with migration")
plt.xlabel("Time")
plt.ylabel("Memory [Gb]")
fname = "../mig.log"
data = [json.loads(line) for line in open(fname, "r")]
plt.plot(get_zone_memory(data, "zone2"), label="zone2")
plt.plot(get_zone_memory(data, "zone3"), label="zone3")
plt.plot(get_zone_memory(data, "zone4"), label="zone4")
plt.plot(get_zone_memory(data, "zone5"), label="zone5")
plt.legend()
plt.show()
# print("Hi")

