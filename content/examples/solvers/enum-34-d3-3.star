def solve(options):
    red_first = [shelf for shelf in permutations(["red", "b2", "b3", "b4"]) if shelf[0] == "red"]
    return match(options, len(red_first))
