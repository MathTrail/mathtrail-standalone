# Two-digit numbers whose digits differ by 3: go through every two-digit
# number and test its digits, taken off with % 10 and // 10.
# Written from the template digits of combinatorics.enumeration.

SMALLEST = 10  # the two-digit numbers, from
LARGEST = 99  # up to and including
APART = 3  # how far the two digits are from each other

def allowed(number):
    return abs(number // 10 - number % 10) == APART

def solve(options):
    return match(options, len([n for n in range(SMALLEST, LARGEST + 1) if allowed(n)]))
