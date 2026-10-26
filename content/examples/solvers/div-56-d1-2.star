def solve(options):
    # Every number up to 20, kept when every run of five whole numbers in a
    # row, from 1 to 100, holds a multiple of it.
    sure = [d for d in range(1, 21) if all([any([(s + i) % d == 0 for i in range(5)]) for s in range(1, 101)])]
    return match(options, max(sure))
