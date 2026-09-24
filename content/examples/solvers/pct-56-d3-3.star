def solve(options):
    answers = set()
    for max_has in range(1, 1001):
        # The 15 extra stickers are 25% of what Max has.
        if max_has * 25 != 15 * 100:
            continue
        lena_has = max_has + 15
        for p in range(0, 101):
            if 15 * 100 == p * lena_has:
                answers.add(p)
    if len(answers) != 1:
        fail("the numbers give %d answers" % len(answers))
    return match(options, "%d%%" % list(answers)[0])
