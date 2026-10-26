STICKERS = 6  # each child has this many, and no two things are alike
BADGES = 4

def things(owner):
    return [(owner, "sticker", i) for i in range(STICKERS)] + [(owner, "badge", i) for i in range(BADGES)]

def solve(options):
    # A swap is one of Ann's things for one of Ben's, both of one kind.
    swaps = [pair for pair in product(things("Ann"), things("Ben")) if pair[0][1] == pair[1][1]]
    return match(options, len(swaps))
