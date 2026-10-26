DIGITS = [1, 2, 3, 4, 5, 6, 7]

def solve(options):
    # Every row of the seven digits; for each, the six counts of digits
    # standing between 1 and 2, 2 and 3 and so on, and how many are odd.
    most = 0
    for row in permutations(DIGITS):
        place = {digit: i for i, digit in enumerate(row)}
        odd = len([d for d in DIGITS[:-1] if (abs(place[d] - place[d + 1]) - 1) % 2 == 1])
        most = max(most, odd)
    return match(options, most)
