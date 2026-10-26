def solve(options):
    # Every way the four could be knights or liars and every set of those of
    # Ann, Ben and Kim who ate the cake, whom the sentences speak of; a
    # knight's sentence is true and a liar's false.
    counts = set()
    for ann, ben, kim, dan in product([True, False], repeat=4):
        for ann_ate, ben_ate, kim_ate in product([True, False], repeat=3):
            said = [ben_ate and kim_ate, not ann_ate and not kim_ate, not ann and not ben, len([x for x in [ann, ben, kim] if x]) == 2]
            if said == [ann, ben, kim, dan]:
                counts.add(len([x for x in [ann, ben, kim, dan] if x]))
    if not counts:
        fail("no way fits")
    if len(counts) != 1:
        return match(options, "It depends on who ate the cake")
    return match(options, list(counts)[0])
