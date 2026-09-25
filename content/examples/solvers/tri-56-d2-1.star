TERMS = 5  # 1, 11, 111, 1111 and 11111

def solve(options):
    # The long way: write each number out as its ones and add them up.
    return match(options, sum([int("1" * length) for length in range(1, TERMS + 1)]))
