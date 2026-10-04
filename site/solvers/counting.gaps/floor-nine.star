# Two girls walk up from floor 1 of one house, one to floor 9 and the other to
# floor 3: count the flights of stairs each one climbs as pairs of neighbouring
# floors, and see how many times as many the first one climbs. The steps say
# it is a whole number of times, so that is checked too.
# Written from the template neighbours of counting.gaps.

FROM = 1  # the floor both start on
LENA = 9  # the floor Lena lives on
MASHA = 3  # the floor Masha lives on

def flights(start, end):
    # One flight between every floor and the next.
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    lena, masha = flights(FROM, LENA), flights(FROM, MASHA)
    if lena % masha != 0:
        fail("%d flights are not a whole number of times %d" % (lena, masha))
    return match(options, lena // masha)
