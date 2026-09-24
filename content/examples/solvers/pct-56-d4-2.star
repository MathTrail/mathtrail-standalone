def solve(options):
    stamps = 200
    mia = []
    for mia_has in range(0, stamps + 1):
        sam_has = stamps - mia_has
        # 20% of Mia's and 30% of Sam's are the same whole number of stamps.
        if mia_has * 20 % 100 == 0 and sam_has * 30 % 100 == 0 and mia_has * 20 == sam_has * 30:
            mia.append(mia_has)
    if len(mia) != 1:
        fail("%d ways of splitting the stamps fit" % len(mia))
    return match(options, mia[0])
