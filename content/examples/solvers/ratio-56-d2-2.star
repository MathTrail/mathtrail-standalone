def solve(options):
    shells = 56
    # Every split of the shells between the boxes, kept when box A to box B is 3 to 4.
    box_a = [a for a in range(0, shells + 1) if a * 4 == (shells - a) * 3]
    if len(box_a) != 1:
        fail("%d splits keep the ratio" % len(box_a))
    return match(options, box_a[0])
