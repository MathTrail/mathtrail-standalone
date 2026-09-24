def solve(options):
    # Nobody likes neither fruit, so the group is apples only, pears only and
    # both; none of those can be more than the 18 who like apples.
    group = set()
    for apples_only, pears_only, both in product(range(19), repeat=3):
        if apples_only + both == 18 and pears_only + both == 14 and both == 7:
            group.add(apples_only + pears_only + both)
    if len(group) != 1:
        fail("the numbers leave %d answers" % len(group))
    return match(options, list(group)[0])
