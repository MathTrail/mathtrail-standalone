NUMBER = 96

def solve(options):
    # Every way to write the number as a product of whole numbers above 1, the
    # factors from the smallest up. Ones then fill the sum up to the number,
    # since they leave the product as it is, and there is always room for them:
    # numbers above 1 never add up to more than they multiply to. A function
    # may not call itself here, so the ways are walked with a stack of their own.
    fewest = None
    ways = [([], NUMBER)]
    while ways:
        factors, left = ways.pop()
        if left == 1:
            count = len(factors) + NUMBER - sum(factors)
            if count >= 2 and (fewest == None or count < fewest):
                fewest = count
            continue
        smallest = factors[-1] if factors else 2
        for factor in range(smallest, left + 1):
            if left % factor == 0:
                ways.append((factors + [factor], left // factor))
    if fewest == None:
        fail("no sum of two or more numbers fits")
    return match(options, fewest)
