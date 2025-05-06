
import random
import sys
import os

random.seed(69)

import shelve
import time
from contextlib import contextmanager
from functools import lru_cache
from time import sleep
import math
import logging
log = logging.getLogger()
from pathlib import Path
import file_cache
import file_cache_test_worker

TRIES = 10
DOMAIN = 4
DELAY = 2.0 # sec
WORKER_ID = None

pref = file_cache_test_worker.pref
value = file_cache_test_worker.value

@lru_cache(maxsize=32)
def find_ratio(a, n, s, r0=None):
    """given the start (a), count (n), and sum (s), find the ratio required"""
    if n == 2:
        return s / a - 1
    an = a * n
    if n == 1 or s == an:
        return 1
    if r0 is None:
        # we only need to know which side of 1 to guess, and the iteration will work
        r0 = 1.1 if an < s else 0.9

    # geometric sum formula
    def f(r):
        return a * (1 - r**n) / (1 - r) - s

    # derivative of f
    def df(r):
        rm1 = r - 1
        rn = r**n
        return (a * (rn * (n * rm1 - r) + r)) / (r * rm1**2)

    MIN_DR = 0.0001  # negligible change
    r = r0
    # print(f"r(0)={r0}")
    MAX_TRIES = 64
    for i in range(1, MAX_TRIES + 1):
        try:
            dr = f(r) / df(r)
        except ZeroDivisionError:
            log.error(f"Failed to find ratio due to zero division! Returning r={r0}")
            return r0
        r = r - dr
        # print(f"r({i})={r}")
        # if the change in r is small, we are close enough
        if abs(dr) < MIN_DR:
            break
    else:
        log.error(f"Could not find ratio after {MAX_TRIES}! Returning r={r0}")
        return r0
    return r


def backoff_delay(start, timeout=None, ratio=None, count: int = 0):
    """generates `count` waits starting at `start`
    sum of waits is `timeout` or each one is `ratio` bigger than the last
    the last wait is always 0"""
    # timeout or ratio must be set but not both
    assert (timeout is None) ^ (ratio is None)
    assert ratio is None or ratio > 0
    assert timeout is None or timeout >= start
    assert (count > 1 or timeout is not None) and isinstance(count, int)
    assert start > 0

    if count == 0:
        # Equation for auto-count is tuned to have a max of
        # ~int(timeout) counts with a start wait of <0.01.
        # Increasing start wait decreases count eg.
        # backoff_delay(10, timeout=60) -> count = 5
        count = int(
            (timeout / ((start + 0.05) ** (1 / 2)) + 2) // math.log(timeout + 2)
        )

    yield start
    # if ratio is set:
    # timeout = start * (1 - ratio**(count - 1)) / (1 - ratio)
    if ratio is None:
        ratio = find_ratio(start, count - 1, timeout)

    wait = start
    # we have start and 0, so we only need to generate count - 2
    for _ in range(count - 2):
        wait *= ratio
        yield wait
    yield 0
    return


template_cache_path = Path("/tmp/test_f_c/template_info.cache")
template_cache_path_exists = Path("/tmp/test_f_c/template_info.cache.exists")

@contextmanager
def template_cache(writeback=False):
    flag = "c" if writeback else "r"
    err = None
    for wait in backoff_delay(0.125, timeout=60, count=20):
        try:
            cache = shelve.open(
                str(template_cache_path), flag=flag, writeback=writeback
            )
            break
        except OSError as e:
            err = e
            log.debug(f"Failed to access template info cache: {e}")
            sleep(wait)
            continue
    else:
        # reached max_count of waits
        raise Exception(f"Failed to access cache file. latest error: {err}")
    try:
        yield cache
    finally:
        cache.close()

def compute(v: int):
    key = str(v)
    if template_cache_path_exists.exists():
        with template_cache() as cache:
            if key in cache:
                res = cache[key]
                assert res == value(v)
                print(f"{pref()} HIT {v}")
                return
                
    print(f"{pref()} MISS {v}")
    time.sleep(DELAY)

     # keep write access open for minimum time
    with template_cache(writeback=True) as cache:
        cache[key] = value(v)
        # Note that shelve database may add custom suffix on top of the path
        # > The filename specified is the base filename for the underlying database. 
        # We find all files which could be the database and open the first one.
        cache_files = [filename for filename in os.listdir(template_cache_path.parent) if filename.startswith(template_cache_path.name)]
        if not cache_files:
            # No cache files found, skip
            return
        # Change ownership of the cache files to the slurm user
        # This is needed to avoid permission issues when running slurmsync
        # as a different user (e.g. when using sudo)
        for filename in cache_files:
            file_cache._chown_slurm(Path(template_cache_path.parent) / filename)

    # Create a marker file to indicate that the cache file exists
    template_cache_path_exists.touch()
    file_cache._chown_slurm(template_cache_path_exists)
    print(f"{pref()} SAVED {v}")
    

if __name__ == "__main__":
    WORKER_ID = sys.argv[1]
    for _ in range(TRIES):
        compute(random.randint(0, DOMAIN))