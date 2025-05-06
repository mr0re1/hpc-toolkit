
import random
import sys

random.seed(69)

import file_cache
import time

TRIES = 10
DOMAIN = 4
DELAY = 2.0 # sec
WORKER_ID = None

def pref():
    ms = int(time.time() * 1000)
    return f"{ms} {WORKER_ID}"

def compute(v: int):
    cache = file_cache.cache("test_f_c")
    if isinstance(cache, file_cache.NoCache):
        print(f"{pref()} NO CACHE")
    key = str(v)
    if (res := cache.get(key)) is not None:
        assert res == v*v, f"{res} != {v*v}"
        print(f"{pref()} HIT {v}")
        return
    print(f"{pref()} MISS {v}")
    time.sleep(DELAY)
    cache.set(key, v*v)
    print(f"{pref()} SAVED {v}")

if __name__ == "__main__":
    WORKER_ID = sys.argv[1]
    for _ in range(TRIES):
        compute(random.randint(0, DOMAIN))