import json
import threading
import time
import os

IP = "192.168.122.31"
PORT = 1323

file_path = "arrivi.json"

def execute_after_delay(delay, value):
    time.sleep(delay)
    f = value[0]
    c = value[1]

    if f == "f1":
        param = 8100
    elif f == "f2":
        param = 6000
    elif f == "f3":
        param = 7200
    elif f == "f4":
        param = 6550
    else:
        param = 8500

    command = f"bin/serverledge-cli invoke -H {IP} -P {PORT} -f {f} -c \"{c}\" -p \"n:{param}\""
    os.system(command)

with open(file_path, "r") as f:
    data = json.load(f)

threads = []
for key, value in data.items():
    delay = float(key)
    thread = threading.Thread(target=execute_after_delay, args=(delay, value))
    threads.append(thread)
    thread.start()

for thread in threads:
    thread.join()

print("Tutti i thread hanno terminato.")
