FIRST = 6
SECOND = 17  # stands exactly opposite FIRST
LARGEST = 2 * SECOND  # the circle is twice the steps between two opposite children, and those are fewer than SECOND

def opposite(children, one, other):
    # Opposite means as many steps one way round the circle as the other.
    steps = (other - one) % children
    return steps * 2 == children

def solve(options):
    found = [n for n in range(SECOND, LARGEST + 1) if opposite(n, FIRST, SECOND)]
    if len(found) != 1:
        fail("%d circles fit" % len(found))
    return match(options, found[0])
