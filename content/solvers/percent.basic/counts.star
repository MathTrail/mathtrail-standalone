# For counts given as percentages of a total: try every total, keep those in
# which each percentage is a whole number of things and everything the
# question says holds, and read the answer off them.
# From reference task pct-56-d4-1.

LARGEST = 1000  # the largest total worth trying

def percent_of(count, percent):
    # A percentage of a count, or None when it is not a whole number.
    if count * percent % 100 != 0:
        return None
    return count * percent // 100

def fits(total):
    # Everything the question says about this total. Here: 25% of the apples
    # are red, and after 6 red ones are eaten, 10% of the apples left are red.
    red = percent_of(total, 25)
    if red == None or red < 6:
        return False
    return (red - 6) * 100 == (total - 6) * 10

def solve(options):
    totals = [total for total in range(1, LARGEST + 1) if fits(total)]
    if len(totals) != 1:
        fail("%d totals fit, and the question has one" % len(totals))
    return match(options, totals[0])  # or the count the question asks about
