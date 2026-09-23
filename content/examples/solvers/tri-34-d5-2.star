def solve(options):
    # The last step of the trick is a division by three, multiplied through.
    fitting = [n for n in range(1001) if (n + 3) * 3 - 3 == 24]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])
