# For ages, "in how many years" or "how old": try every whole number of years
# and keep those the words allow. Exactly one must be left; any other count
# means the words do not settle the question.
# From reference task cal-12-d4-4.

CHILD_NOW = 5  # Ben's age now
PARENT_NOW = 29  # Mum's age now
TIMES = 3  # how many times as old as Ben Mum is to be

def fits(years):
    return PARENT_NOW + years == TIMES * (CHILD_NOW + years)

def solve(options):
    fitting = [years for years in range(1, 100) if fits(years)]
    if len(fitting) != 1:
        fail("%d numbers of years fit the clue" % len(fitting))
    return match(options, fitting[0])
