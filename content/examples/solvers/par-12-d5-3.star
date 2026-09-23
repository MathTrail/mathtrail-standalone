def solve(options):
    # Rubbing out two numbers and writing their sum leaves the total on the
    # board as it was, so the last number left is that total, however the
    # pairs are chosen. The prototype sampled thirty games to show it; the
    # invariant is the proof itself.
    last = sum(range(1, 21))
    return match(options, "even" if last % 2 == 0 else "odd")
