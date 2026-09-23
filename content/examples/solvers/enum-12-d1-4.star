def solve(options):
    first = [line for line in permutations(["Leo", "Mia", "Zoe"]) if line[0] == "Leo"]
    return match(options, len(first))
