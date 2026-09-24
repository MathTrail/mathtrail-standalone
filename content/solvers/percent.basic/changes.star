# For a price that changes by percentages, each of the price at that moment:
# try every starting price, make the changes in turn in hundredths, and keep
# the prices in which every change is a whole number of coins and the last
# price is the one the question gives.
# From reference task pct-56-d3-1.

CHANGES = [-20]  # each change in per cent, in order: positive for a rise, negative for a fall
END = 60  # the price after the last change
LARGEST = 1000  # the largest starting price worth trying

def after(price):
    # The price after every change, or None when a change is not a whole
    # number of coins.
    for change in CHANGES:
        if price * change % 100 != 0:
            return None
        price += price * change // 100
    return price

def solve(options):
    # When the question gives the starting price and asks the last one, match
    # after(that price) instead.
    starts = [price for price in range(1, LARGEST + 1) if after(price) == END]
    if len(starts) != 1:
        fail("%d starting prices end at %d, and the question has one" % (len(starts), END))
    return match(options, starts[0])
