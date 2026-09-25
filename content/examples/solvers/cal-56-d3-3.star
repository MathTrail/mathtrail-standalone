TOGETHER = 110  # the three ages now
DAD_AT_BIRTH = 26  # Dad's age when the son was born
GRANDPA_AT_BIRTH = 51

def fits(son):
    return son + (son + DAD_AT_BIRTH) + (son + GRANDPA_AT_BIRTH) == TOGETHER

def solve(options):
    fitting = [son for son in range(120) if fits(son)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
