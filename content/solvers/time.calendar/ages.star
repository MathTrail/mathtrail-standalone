# For ages, "how old" or "in how many years": try every whole number for the
# one the words leave unknown, here Ben's age now, and keep those the words
# allow. Exactly one must be left; any other count means the words do not
# settle the question.
# From reference task cal-12-d3-2.

YEARS_ON = 5  # how many years from now Ben's age is given
BEN_THEN = 12  # Ben's age then
OLDER_BY = 3  # how many years older than Ben Ann is

def fits(ben_now):
    return ben_now + YEARS_ON == BEN_THEN

def solve(options):
    fitting = [age for age in range(100) if fits(age)]
    if len(fitting) != 1:
        fail("%d ages of Ben fit the clue" % len(fitting))
    return match(options, fitting[0] + OLDER_BY)
