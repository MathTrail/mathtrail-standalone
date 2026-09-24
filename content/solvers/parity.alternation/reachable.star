# For "which of these can it be": make every result the moves can reach, and
# keep the numbers of the options that are among them. Exactly one must be.
# From reference task par-34-d3-4.

JUMPS = 9  # jumps the grasshopper makes, each one step left or right
CANDIDATES = [0, 2, 4, 7, 10]  # the numbers the options name

def solve(options):
    ends = set([sum(jumps) for jumps in product([1, -1], repeat=JUMPS)])
    fitting = [n for n in CANDIDATES if n in ends]
    if len(fitting) != 1:
        fail("%d of the numbers can be reached" % len(fitting))
    return match(options, fitting[0])
