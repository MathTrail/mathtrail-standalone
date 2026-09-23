def solve(options):
    fitting = [petya for petya in range(1, 50) if petya + 4 * petya == 40]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
