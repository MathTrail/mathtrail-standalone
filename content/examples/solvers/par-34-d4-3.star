def solve(options):
    # Rubbing out two numbers and writing their difference leaves the total on
    # the board even or odd, as it was: x + y and x - y differ by 2y. So the
    # last number has the parity of 1 + 2 + ... + 20, however the steps go.
    # The prototype sampled twenty thousand games to show it; the invariant is
    # the proof itself, and it has to single out one option on its own.
    parity = sum(range(1, 21)) % 2
    fitting = [n for n in [1, 2, 3, 5, 7] if n % 2 == parity]
    if len(fitting) != 1:
        fail("%d of the numbers have the right parity" % len(fitting))
    return match(options, fitting[0])
