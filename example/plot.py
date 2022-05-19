import json
import matplotlib.pyplot as plt

# with open("./mig.log", "r") as f:

fname = "../nomig.log"
data = [json.loads(line) for line in open(fname, "r")]

# print(data[0])
z2_mem = []
z3_mem = []


def get_memory(t, node):
    z2 = t["Nodes"][node]["TotalResourceUsage"]["memory"]
    z2 = int(z2)
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

    return bytes / d


for t in data:
    z2 = get_memory(t, "zone2")
    # print(z2)
    d = 1 << 20

    # print("CONV", z2 / d)
    # print(z2)
    z2_mem.append(bytesto(z2, "g"))
    z3 = get_memory(t, "zone3")
    z3_mem.append(bytesto(z3, "g"))

# d = 1 << 20
# print(bytesto(202849602216, "g"), "\n", 202849602216 / d)
plt.plot(z2_mem)
plt.plot(z3_mem)
plt.show()
# print("Hi")

