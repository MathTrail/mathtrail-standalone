def solve(options):
    fitting = [now for now in range(50) if now + 5 == 12]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0] + 3)
