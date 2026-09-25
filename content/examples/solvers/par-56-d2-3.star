NUMBERS = 20
VALUES = [1, 2, 3, 4]  # whole numbers to try, two odd and two even
CANDIDATES = [10, 8, 20, 12, 7]  # how many even numbers the options name

def solve(options):
    # Every way to pick the numbers one after another from VALUES, kept as how
    # many of them are even and what they add up to. An option no pick reaches
    # is out of reach with these values only; that larger numbers change
    # nothing is what the parity of a sum says.
    reached = set([(0, 0)])
    for _ in range(NUMBERS):
        reached = set([(evens + (1 if value % 2 == 0 else 0), total + value) for evens, total in reached for value in VALUES])
    fitting = [n for n in CANDIDATES if any([evens == n and total % 2 == 1 for evens, total in reached])]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
