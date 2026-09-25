def solve(options):
    found = []
    for x, y, z in permutations(range(10), 3):  # different letters, different digits
        xyz = 100 * x + 10 * y + z
        if x != 0 and xyz * 3 == 111 * z:  # a number does not start with 0
            found.append(xyz)
    if len(found) != 1:
        fail("%d numbers fit" % len(found))
    return match(options, found[0])
