PATH = 300  # metres
REACH = 25  # metres the scent goes from a bush, either way

def solve(options):
    # A search over where the bushes stand, one bush at a time, in whole
    # metres along the path. The scent covers the path from its start up to
    # some point, and the next bush may stand anywhere its scent still joins
    # on to what is covered, leaving no gap. The first number of bushes whose
    # scent can reach the end of the path is the answer.
    seen = {0: True}
    frontier = [0]  # how far the scent can reach with the bushes so far
    bushes = 0
    while frontier:
        bushes += 1
        ahead = []
        for reached in frontier:
            for stands in range(max(0, reached - REACH), min(PATH, reached + REACH) + 1):
                new = min(PATH, stands + REACH)
                if new == PATH:
                    return match(options, bushes)
                if new > reached and new not in seen:
                    seen[new] = True
                    ahead.append(new)
        frontier = ahead
    fail("the scent never reaches the end of the path")
