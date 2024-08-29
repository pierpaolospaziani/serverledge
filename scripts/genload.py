import numpy as np
import random
import threading
import time
import os

IP = "192.168.122.31"
PORT = 1323

seed_value = 12345
np.random.seed(seed_value)
random.seed(seed_value)

func_to_list = {"f1":[], "f2":[], "f3":[], "f4":[], "f5":[]}

classes = [
    {"name": "standard", "arrival_weight": 0.7},
    {"name": "critical-1", "arrival_weight": 0.1},
    {"name": "critical-2", "arrival_weight": 0.1},
    {"name": "batch", "arrival_weight": 0.1}
]

functions = [
    {"name": "f1", "rate": 0.8, "param": 8100},
    {"name": "f2", "rate": 1.6, "param": 6000},
    {"name": "f3", "rate": 4.2, "param": 7200},
    {"name": "f4", "rate": 0.6, "param": 6550},
    {"name": "f5", "rate": 1.4, "param": 8500}
]

weights = [cls['arrival_weight'] for cls in classes]
cumulative_weights = np.cumsum(weights)

def generate_poisson_arrivals(rate, duration):
    return np.cumsum(np.random.exponential(1/rate, int(rate * duration)))

def select_class():
    rand_value = random.uniform(0, cumulative_weights[-1])
    for i, weight in enumerate(cumulative_weights):
        if rand_value <= weight:
            return classes[i]['name']


class ArrivalGenerator(threading.Thread):
    def __init__(self, function, duration):
        super().__init__()
        self.function = function
        self.duration = duration

    def run(self):
        arrivals = generate_poisson_arrivals(self.function['rate'], self.duration)
        for arrival_time in arrivals:
            class_name = select_class()
        
            func_to_list[self.function['name']].append((arrival_time, self.function['name'], self.function['param'], class_name))


def invoke_function(function_name, param, class_name):
    command = f"bin/serverledge-cli invoke -H {IP} -P {PORT} -f {function_name} -c \"{class_name}\" -p \"n:{param}\""
    os.system(command)


duration = 3600

threads = []
for func in functions:
    thread = ArrivalGenerator(func, duration)
    threads.append(thread)
    thread.start()

for thread in threads:
    thread.join()

complete_list = []
for key in func_to_list.keys():
    for t in func_to_list[key]:
        complete_list.append(t)
complete_list.sort(key=lambda x: x[0])

for i, (arrival_time, function_name, param, class_name) in enumerate(complete_list):
    if i == 0:
        delay = arrival_time
    else:
        delay = arrival_time - complete_list[i-1][0]
    time.sleep(delay)
    threading.Thread(target=invoke_function, args=(function_name, param, class_name)).start()