def solve(options):
    fitting = [times for times in range(1, 20) if 9 * times == 36]
    if len(fitting) != 1:
        fail("%d numbers fit the clue" % len(fitting))
    return match(options, fitting[0])
