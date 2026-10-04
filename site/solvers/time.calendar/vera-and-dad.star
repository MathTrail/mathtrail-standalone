# Dad is 28 years older than Vera, and in 4 years he will be 3 times as old as
# she will be: try every whole age for Vera now and keep those the words
# allow. Exactly one must be left; any other count means the words do not
# settle the question.
# Written from the template ages of time.calendar.

OLDER = 28  # years Dad is older than Vera
LATER = 4  # years from now
TIMES = 3  # how many times as old as Vera Dad will be then

def fits(vera):
    return vera + OLDER + LATER == TIMES * (vera + LATER)

def solve(options):
    fitting = [vera for vera in range(1, 100) if fits(vera)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
