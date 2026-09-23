def gaps(objects):
    return len([(i, i + 1) for i in range(objects - 1)])

def solve(options):
    beads = "R" + "BBR" * gaps(5)
    if beads.count("R") != 5:
        fail("the string does not hold five red beads")
    return match(options, beads.count("B"))
