# A flag of two stripes, one above the other, each one of three colours, the
# two of different colours: make every flag and count them.
# Written from the template arrangements of combinatorics.enumeration.

COLOURS = ["red", "yellow", "green"]

def allowed(flag):
    top, bottom = flag
    return top != bottom  # the two stripes are of different colours

def solve(options):
    return match(options, len([flag for flag in product(COLOURS, repeat=2) if allowed(flag)]))
