def solve(options):
    # Every four-digit number of four different odd digits that divides by 11.
    found = []
    for n in range(1000, 10000):
        digits = [n // 1000, n // 100 % 10, n // 10 % 10, n % 10]
        if len(set(digits)) == 4 and all([d % 2 == 1 for d in digits]) and n % 11 == 0:
            found.append(n)
    if not found:
        fail("no number fits")
    return match(options, max(found))
