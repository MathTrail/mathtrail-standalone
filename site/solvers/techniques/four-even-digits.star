# How many two-digit numbers the given digits make, no digit used twice: go
# through every two-digit number and test its digits, taken off one at a time
# with % 10 and // 10.
# Written from the template digits of combinatorics.enumeration.

ALLOWED = [2, 4, 6, 8]  # the digits the numbers may use

def digits(number):
    found = []
    while number > 0:
        found.append(number % 10)
        number = number // 10
    return found

def allowed(number):
    found = digits(number)
    return all([digit in ALLOWED for digit in found]) and len(set(found)) == len(found)

def solve(options):
    return match(options, len([n for n in range(10, 100) if allowed(n)]))
