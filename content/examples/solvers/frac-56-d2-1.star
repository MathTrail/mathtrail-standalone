def solve(options):
    # Every number whose two thirds are a whole 24.
    numbers = [n for n in range(1, 1001) if n % 3 == 0 and n // 3 * 2 == 24]
    if len(numbers) != 1:
        fail("%d numbers have two thirds equal to 24" % len(numbers))
    number = numbers[0]
    if number % 4 != 0:
        fail("three quarters of %d is not a whole number" % number)
    return match(options, number // 4 * 3)
