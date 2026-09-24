# For the smallest or the largest number with a property, or how many numbers
# have it: try every number in a range and test it.
# From reference task div-56-d3-2.

SMALLEST = 2  # the first number to try: here, more than one lemon
LARGEST = 10000  # the last number to try
CAPPED = True  # True when LARGEST only stops the search, False when the question itself sets it

def has_property(n):
    # Here: counted out in twos, threes, fours, fives or sixes, one is left over.
    return all([n % group == 1 for group in [2, 3, 4, 5, 6]])

def largest(found):
    # The largest number found. When LARGEST only stops the search, a larger
    # one could lie beyond it, so the answer is trusted only when the search
    # went on at least as far again past it and found nothing more; a limit
    # the question sets is an answer like any other.
    if CAPPED and found[-1] * 2 > LARGEST:
        fail("the largest number found is too near where the search stops; search further")
    return found[-1]

def solve(options):
    found = [n for n in range(SMALLEST, LARGEST + 1) if has_property(n)]
    if len(found) == 0:
        fail("no number from %d to %d has the property" % (SMALLEST, LARGEST))
    return match(options, found[0])  # largest(found) for the largest, len(found) for how many
