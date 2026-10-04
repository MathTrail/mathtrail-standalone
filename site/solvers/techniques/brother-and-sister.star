# Two ages that differ by a known number and add up to another: try every age
# of the younger one and keep those the words allow. Exactly one must be left;
# any other count means the words do not settle the question.
# Written from the template ages of time.calendar.

OLDER_BY = 4  # years the brother is older than his sister
TOGETHER = 16  # years the two are together

def fits(sister):
    return sister + (sister + OLDER_BY) == TOGETHER

def solve(options):
    fitting = [sister for sister in range(TOGETHER + 1) if fits(sister)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0] + OLDER_BY)  # the brother's age
