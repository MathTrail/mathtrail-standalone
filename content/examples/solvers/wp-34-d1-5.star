def solve(options):
    # Which weights go opposite the sugar, and which go on the pan with it.
    placements = {
        "Both weights on the pan opposite the sugar": ([1, 2], []),
        "Both weights on the pan with the sugar": ([], [1, 2]),
        "The 1 kg weight with the sugar, the 2 kg weight opposite": ([2], [1]),
        "The 2 kg weight with the sugar, the 1 kg weight opposite": ([1], [2]),
    }
    working = [label for label in placements if sum(placements[label][0]) - sum(placements[label][1]) == 3]
    if len(working) > 1:
        fail("%d of the placements weigh out 3 kg" % len(working))
    return match(options, working[0] if working else "It cannot be done in one weighing")
