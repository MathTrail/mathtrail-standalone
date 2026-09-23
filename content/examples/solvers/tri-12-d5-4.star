def solve(options):
    # Half of the sweets has to be a whole number of sweets.
    fitting = [n for n in range(1001) if n % 2 == 0 and n - n // 2 - 1 == 4]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])
