def solve(options):
    return match(options, len(set(permutations(["R", "R", "R", "U", "U"]))))
