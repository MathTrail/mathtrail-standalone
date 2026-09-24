def solve(options):
    pupils = 30
    neither = []
    for maths_only, art_only, both in product(range(pupils + 1), repeat=3):
        rest = pupils - maths_only - art_only - both
        if rest >= 0 and maths_only + both == 18 and art_only + both == 21:
            neither.append(rest)
    return match(options, max(neither))
