import time
import random

SLEEP = 0
FAIL_RATE = 0.5

if SLEEP:
    time.sleep(SLEEP)
assert random.random() > FAIL_RATE, "bad luck"
