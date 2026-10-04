# A bus route of 6 stops, with the same minutes from every stop to the next:
# count the gaps the bus drives through as pairs of neighbouring stops, and
# give each gap its minutes.
# Written from the template neighbours of counting.gaps.

STOPS = 6  # stops on the route, the first and the last among them
MINUTES = 3  # minutes from one stop to the next

def gaps(stops):
    # One gap between every stop and the next.
    return len([(stop, stop + 1) for stop in range(1, stops)])

def solve(options):
    return match(options, gaps(STOPS) * MINUTES)
