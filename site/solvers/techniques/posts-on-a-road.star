# Posts along a road, one every few metres, both ends included: put every post
# at its place and count the places, instead of adding one by hand.
# Written from the template positions of counting.gaps.

ROAD = 100  # metres from one end of the road to the other
APART = 5  # metres from one post to the next

def solve(options):
    return match(options, len([metre for metre in range(0, ROAD + 1, APART)]))
