def solve(options):
    # Every whole number of squares each child could have eaten: three
    # quarters of Mia's 20 and two thirds of Leo's 24.
    mia = [n for n in range(0, 21) if n * 4 == 20 * 3]
    leo = [n for n in range(0, 25) if n * 3 == 24 * 2]
    if len(mia) != 1 or len(leo) != 1 or mia[0] == leo[0]:
        fail("the bars do not give one child who ate more")
    if mia[0] > leo[0]:
        return match(options, "Mia, %d" % mia[0])
    return match(options, "Leo, %d" % leo[0])
