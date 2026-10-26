def solve(options):
    # Every two-digit number, kept when the product of its digits is not 0
    # and divides it.
    found = []
    for n in range(10, 100):
        product = (n // 10) * (n % 10)
        if product != 0 and n % product == 0:
            found.append(n)
    return match(options, len(found))
