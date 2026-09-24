def solve(options):
    # k/12 is bigger than 1/3 when k * 3 > 12 * 1, and smaller than 3/4 when
    # k * 4 < 12 * 3: fractions compared as pairs of whole numbers.
    between = [k for k in range(1, 13) if k * 3 > 12 * 1 and k * 4 < 12 * 3]
    return match(options, len(between))
