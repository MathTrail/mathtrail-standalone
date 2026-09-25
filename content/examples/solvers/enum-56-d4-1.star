SMALLEST = 100  # the three-digit numbers, from
LARGEST = 999  # up to and including
USABLE = [0, 1, 2, 3, 4]  # the digits a number may use

def digits(number):
    found = []
    while number > 0:
        found.append(number % 10)
        number = number // 10
    return found

def allowed(number):
    written = digits(number)
    usable = all([digit in USABLE for digit in written])
    different = len(set(written)) == len(written)
    return number % 2 == 0 and usable and different

def solve(options):
    return match(options, len([n for n in range(SMALLEST, LARGEST + 1) if allowed(n)]))
