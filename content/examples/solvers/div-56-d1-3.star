def solve(options):
    # Every pair of numbers up to 200 with remainders 5 and 7 when divided by 9.
    remainders = set()
    for a in range(0, 201):
        if a % 9 != 5:
            continue
        for b in range(0, 201):
            if b % 9 == 7:
                remainders.add((a + b) % 9)
    if len(remainders) != 1:
        fail("%d remainders fit" % len(remainders))
    return match(options, list(remainders)[0])
