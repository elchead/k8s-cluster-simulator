import json

# import matplotlib.pyplot as plt
export_filepath = "/Users/I545428/gh/controller-simulator/pods.json"

id_to_pod = {}
with open("./example/entities_12:5-8-12.json", "r") as f:
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

# print(id_to_pod)

not_found_ids = []
with open("./example/pod_500_12:5-8-12.json", "r") as f:
    res = json.load(f)
    data = res["result"][0]["data"]
    pods = []
    for d in data:
        container = d["dimensions"][0]
        memory = d["values"]
        time = d["timestamps"]

        # filter nil data
        time = [t for idx, t in enumerate(time) if memory[idx] != None]
        memory = [t for t in memory if t != None]
        try:
            podname = id_to_pod[container]
        except KeyError:
            # print(memory)
            not_found_ids.append(container)

        pod = {"Name": podname, "Memory": memory, "Time": time}
        pods.append(pod)

    print("Found", len(pods), "pods")
    with open(export_filepath, "w") as outfile:
        json.dump(pods, outfile)
print(not_found_ids)
# plt.plot(v)
# plt.show()
