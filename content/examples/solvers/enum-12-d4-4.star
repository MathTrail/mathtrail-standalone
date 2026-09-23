def solve(options):
    apart = []
    for row in permutations(["Ann", "Ben", "Kim"]):
        place = {}
        for index, child in enumerate(row):
            place[child] = index
        if abs(place["Ann"] - place["Ben"]) != 1:
            apart.append(row)
    return match(options, len(apart))
