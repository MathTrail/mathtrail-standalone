def split_weighings(total, goal, limit):
    # Everything is counted in 64ths of a kilogram, so that halving a pile
    # stays a whole number for as many weighings as the search allows: the
    # exact fractions of the prototype are exact integers here.
    scale = 64
    level = set([(total * scale,)])
    for weighings in range(limit + 1):
        # Piles poured together make the goal without weighing anything more.
        for piles in level:
            for size in range(1, len(piles) + 1):
                for group in combinations(piles, size):
                    if sum(group) == goal * scale:
                        return weighings
        # One weighing more: pour any piles together and split them in half.
        following = set()
        for piles in level:
            for mask in range(1, 1 << len(piles)):
                chosen = [piles[i] for i in range(len(piles)) if (mask >> i) & 1 == 1]
                rest = [piles[i] for i in range(len(piles)) if (mask >> i) & 1 == 0]
                if sum(chosen) % 2 != 0:
                    fail("a pile no longer halves into whole 64ths of a kilogram")
                half = sum(chosen) // 2
                following.add(tuple(sorted(rest + [half, half])))
        level = following
    return None

def solve(options):
    weighings = split_weighings(24, 9, 6)
    return match(options, "It is impossible" if weighings == None else weighings)
