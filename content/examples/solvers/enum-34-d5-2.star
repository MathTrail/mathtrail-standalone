def solve(options):
    choices = [captain for team in combinations(range(7), 3) for captain in team]
    return match(options, len(choices))
