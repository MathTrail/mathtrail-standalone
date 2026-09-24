def can_score(total):
    # Some darts score 8 each, and the rest of the total has to come in threes.
    return any([(total - 8 * eights) % 3 == 0 for eights in range(0, total // 8 + 1)])

def solve(options):
    limit = 100
    missing = [total for total in range(1, limit + 1) if not can_score(total)]
    largest = max(missing)
    # Past three totals in a row that can be scored, adding threes reaches every
    # larger one; an answer near the limit would mean the search stopped too soon.
    if largest > limit - 3:
        fail("the largest total that cannot be scored is too near %d" % limit)
    return match(options, largest)
