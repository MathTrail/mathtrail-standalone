# A school has 300 pupils, and 35% of them go to a sports club: take 65% of
# the pupils at once, as the steps do, which must be a whole number. The note
# takes 35% and counts those left, so that way is checked to give the same.
# Written from the template counts of percent.basic.

TOTAL = 300  # the pupils of the school
CLUB = 35  # the percentage that goes to a sports club

def percent_of(count, percent):
    # A percentage of a count, or None when it is not a whole number.
    if count * percent % 100 != 0:
        return None
    return count * percent // 100

def solve(options):
    others = percent_of(TOTAL, 100 - CLUB)
    if others == None:
        fail("%d%% of %d is not a whole number of pupils" % (100 - CLUB, TOTAL))
    club = percent_of(TOTAL, CLUB)
    if club == None or TOTAL - club != others:
        fail("taking %d%% and counting those left does not give the same pupils" % CLUB)
    return match(options, others)
