def solve(options):
    cards = range(1, 21)
    target = 21
    # Nothing tells one card of a couple from the other, so hands are counted
    # by couple the way a stock of balls is counted by colour: a hand with no
    # couple holds at most one card of each, and every card with no partner.
    paired = 0
    lone = 0
    for card in cards:
        partner = target - card
        if partner != card and partner in cards:
            paired += 1
        else:
            lone += 1
    couples = paired // 2

    failed = set()
    for from_couples in range(couples + 1):
        for from_lone in range(lone + 1):
            failed.add(from_couples + from_lone)
    for size in range(len(cards) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of cards is ever enough")
