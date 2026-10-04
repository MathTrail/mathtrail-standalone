# 200 kg of watermelons are 99% water, and a week later they are 98% water:
# what is not water stays as it was, so try every weight and keep those of
# which that part is 2%. Exactly one must be left.
# Written from the template counts of percent.basic.

BEFORE = 200  # kilograms at first
WATER_BEFORE = 99  # per cent of water at first
WATER_AFTER = 98  # per cent of water a week later
LARGEST = 1000  # the heaviest weight worth trying

def percent_of(count, percent):
    # A percentage of a count, or None when it is not a whole number.
    if count * percent % 100 != 0:
        return None
    return count * percent // 100

def solve(options):
    flesh = percent_of(BEFORE, 100 - WATER_BEFORE)  # the part that is not water
    if flesh != 2:
        fail("what is not water is %r kg, and the steps say 2 kg" % flesh)
    weights = [weight for weight in range(1, LARGEST + 1) if percent_of(weight, 100 - WATER_AFTER) == flesh]
    if len(weights) != 1:
        fail("%d weights fit, and the question has one" % len(weights))
    return match(options, weights[0])
