def solve(options):
    # The numbers that make 12 whole groups of 7 with 5 left over.
    numbers = [n for n in range(0, 1001) if n // 7 == 12 and n % 7 == 5]
    if len(numbers) != 1:
        fail("%d numbers fit" % len(numbers))
    return match(options, numbers[0])
