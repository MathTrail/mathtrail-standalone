def solve(options):
    start = 12
    # wins[n] says whether the player about to move with n on the board can
    # force a win. From 1 there is no move: no divisor of 1 is smaller than 1.
    wins = [False] * (start + 1)
    for n in range(1, start + 1):
        wins[n] = any([not wins[n - divisor] for divisor in range(1, n) if n % divisor == 0])
    first = [divisor for divisor in range(1, start) if start % divisor == 0 and not wins[start - divisor]]
    return match(options, len(first))
