def solve(options):
    landings = list(range(1, 9 + 1, 2))
    if landings[-1] != 9:
        fail("the last landing is not the ninth floor")
    return match(options, len(landings) - 1)
