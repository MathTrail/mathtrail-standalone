def kind(a, b):
    odd = a % 2 + b % 2
    return {2: "Both are odd", 0: "Both are even", 1: "One is even and one is odd"}[odd]

def solve(options):
    kinds = set([kind(a, b) for a in range(1, 21) for b in range(1, 21) if (a * b) % 2 == 1])
    if len(kinds) != 1:
        fail("%d kinds of pair give an odd product" % len(kinds))
    return match(options, list(kinds)[0])
