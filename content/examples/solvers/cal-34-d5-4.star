def solve(options):
    fitting = [youngest + 3 for youngest in range(50) if sum([youngest + k for k in range(4)]) == 42]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
