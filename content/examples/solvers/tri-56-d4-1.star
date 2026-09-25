LAST = 100  # the numbers from 1 to this

def digits(number):
    found = []
    while number > 0:
        found.append(number % 10)
        number = number // 10
    return found

def solve(options):
    # The long way: every digit of every number.
    return match(options, sum([sum(digits(n)) for n in range(1, LAST + 1)]))
