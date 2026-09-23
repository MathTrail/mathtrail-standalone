def solve(options):
    at_an_end = [line for line in permutations(["Tom", "c1", "c2", "c3"]) if line[0] == "Tom" or line[3] == "Tom"]
    return match(options, len(at_an_end))
