def solve(options):
    # The smallest number above 1 that leaves 1 over in twos, threes, fours, fives and sixes.
    for lemons in range(2, 10001):
        if all([lemons % group == 1 for group in [2, 3, 4, 5, 6]]):
            return match(options, lemons)
    fail("no crate of up to 10000 lemons fits")
