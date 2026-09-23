def trailing_zeros(number):
    zeros = 0
    while number % 10 == 0:
        zeros += 1
        number = number // 10
    return zeros

def solve(options):
    return match(options, trailing_zeros(prod(range(1, 16))))
