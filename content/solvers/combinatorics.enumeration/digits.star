# For "how many numbers" with a property of their digits: go through every
# number of the range and test its digits, taken off one at a time with % 10
# and // 10.
# From reference task enum-34-d5-1.

SMALLEST = 100  # the three-digit numbers, from
LARGEST = 999  # up to and including
DIGIT_SUM = 5  # what the digits add up to

def digits(number):
    found = []
    while number > 0:
        found.append(number % 10)
        number = number // 10
    return found

def allowed(number):
    return sum(digits(number)) == DIGIT_SUM

def solve(options):
    return match(options, len([n for n in range(SMALLEST, LARGEST + 1) if allowed(n)]))
