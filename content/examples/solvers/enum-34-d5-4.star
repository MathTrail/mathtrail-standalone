def solve(options):
    for people in range(2, 100):
        if len(combinations(range(people), 2)) == 28:
            return match(options, people)
    return match(options, "It is impossible")
