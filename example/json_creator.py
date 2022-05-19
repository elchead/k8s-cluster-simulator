import json

# import matplotlib.pyplot as plt

id_to_pod = {}
with open("./example/entities_pods.json", "r") as f:
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
with open("./example/pod_500.json", "r") as f:
    res = json.load(f)
    data = res["result"][0]["data"]
    pods = []
    for d in data:
        container = d["dimensions"][0]
        memory = d["values"]
        time = d["timestamps"]
        try:
            podname = id_to_pod[container]
        except KeyError:
            # print(memory)
            not_found_ids.append(container)

        pod = {"Name": podname, "Memory": memory, "Time": time}
        pods.append(pod)

    with open("./example/pods.json", "w") as outfile:
        json.dump(pods, outfile)
print(not_found_ids)
# plt.plot(v)
# plt.show()
