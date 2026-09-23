def solve(options):
    fitting = [boy for boy in range(1, 65) if boy + 12 * boy == 65]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
