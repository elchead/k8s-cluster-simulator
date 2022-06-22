import json

# import matplotlib.pyplot as plt
export_filepath = (
    "/Users/I545428/gh/controller-simulator/pods_760.json"  # "/Users/I545428/gh/controller-simulator/pods_2715.json"
)
import_filepath = "/Users/I545428/gh/controller-simulator/9-6-2000-9-6-2300_760.json"
# import_filepath = "/Users/I545428/gh/controller-simulator/13-6-1630-14-6-300_2715_2m.json"
entities_path = "/Users/I545428/gh/controller-simulator/entities_9-6-760.json"  # "/Users/I545428/gh/controller-simulator/entities_2715.json"
id_to_pod = {}

# entities
with open(entities_path, "r") as f:
    res = json.load(f)
    ents = res["entities"]
    for e in ents:
        key = e["entityId"]
        name = e["displayName"]
        # print(name)
        n = name.split(" ")
        if len(n) == 1:
            continue
        container = n[1]
        if container == "worker":
            pod = n[0]
            id_to_pod[key] = pod

print(len(id_to_pod), "entities")


# data
not_found_ids = []
with open(import_filepath, "r") as f:
    res = json.load(f)
    data = res["result"][0]["data"]
    pods = []
    for d in data:
        container = d["dimensions"][0]
        memory = d["values"]
        time = d["timestamps"]

        # filter nil data
        time = [t for idx, t in enumerate(time) if memory[idx] != None]
        memory = [int(t) for t in memory if t != None]
        try:
            podname = id_to_pod[container]
            pod = {"Name": podname, "Memory": memory, "Time": time}
            pods.append(pod)
        except KeyError:
            # print(memory)
            not_found_ids.append(container)

    print("Found", len(pods), "pods")
    with open(export_filepath, "w") as outfile:
        json.dump(pods, outfile)
print(not_found_ids)
# plt.plot(v)
# plt.show()
