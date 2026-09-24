def solve(options):
    walnuts = 90
    ratio = [3, 5, 7]
    # Try every size of one part: the baskets hold that part times the ratio,
    # and the sizes that fill them with exactly 90 walnuts are kept.
    answers = set()
    for part in range(1, walnuts + 1):
        baskets = [part * share for share in ratio]
        if sum(baskets) == walnuts:
            answers.add(baskets[2] - baskets[0])
    if len(answers) != 1:
        fail("the baskets can be filled in %d ways" % len(answers))
    return match(options, list(answers)[0])
