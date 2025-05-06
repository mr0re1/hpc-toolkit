from concurrent.futures import ThreadPoolExecutor
import subprocess
import os

CONCUR = 3
WORKERS = 80

def demote(): # as orlov
    os.setgid(89939)
    os.setuid(445305)

def spawn(id):
    cmd = f"/usr/local/google/home/orlov/p310/bin/python ./file_cache_test_worker.py {id}"
    subprocess.run(cmd, check=True, shell=True, preexec_fn=[None, demote][id % 2])


with ThreadPoolExecutor(max_workers=CONCUR) as executor:
    executor.map(spawn, range(WORKERS))