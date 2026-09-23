def solve(options):
    # Every way the games could have gone, kept by how many each player has
    # played: nothing else in the question tells two of them apart.
    games = combinations(range(6), 2)
    reached = {(0, 0, 0, 0, 0, 0): True}
    for game in games:
        grown = {}
        for played in reached:
            grown[played] = True
            after = list(played)
            after[game[0]] += 1
            after[game[1]] += 1
            grown[tuple(after)] = True
        reached = grown
    for played in reached:
        if len(set(played)) == 6:
            return match(options, "No, never")
    return match(options, "Yes, always")
