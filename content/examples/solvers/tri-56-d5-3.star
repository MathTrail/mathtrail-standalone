CARDS = ["1", "3", "7", "8"]

def solve(options):
    # Every number the four cards make, in order, and the largest step between
    # two numbers next to each other.
    numbers = sorted([int("".join(order)) for order in permutations(CARDS)])
    return match(options, max([numbers[i + 1] - numbers[i] for i in range(len(numbers) - 1)]))
