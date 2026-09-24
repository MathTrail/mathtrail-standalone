def solve(options):
    # The eggs grow in step with the people: eggs to people is 9 to 6.
    eggs = [e for e in range(0, 1001) if e * 6 == 9 * 20]
    if len(eggs) != 1:
        fail("%d numbers of eggs keep the recipe's ratio" % len(eggs))
    return match(options, eggs[0])
