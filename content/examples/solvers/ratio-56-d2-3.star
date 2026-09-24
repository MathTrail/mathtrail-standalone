def solve(options):
    silver = 20
    # 3 gold for every 5 silver, and every fish is one or the other.
    gold = [g for g in range(0, 1001) if g * 5 == silver * 3]
    if len(gold) != 1:
        fail("%d numbers of gold fish keep the ratio" % len(gold))
    return match(options, silver + gold[0])
