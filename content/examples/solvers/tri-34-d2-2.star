def solve(options):
    # A quarter of the number plus six is eleven, multiplied through by four.
    fitting = [n for n in range(1001) if n + 24 == 44]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])
