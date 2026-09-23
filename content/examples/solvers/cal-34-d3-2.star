def solve(options):
    fitting = [brother for brother in range(1, 50) if 3 * brother + 4 == 2 * (brother + 4)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, 3 * fitting[0])
