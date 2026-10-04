# Numbers written with the digits 1 and 2 alone, with no 2 right next to
# another 2: make every such number of each length and count them. The
# example's steps build the counts up from the shorter lengths, 2, 3, 5 and
# 8, so those are checked as well.
# Written from the template arrangements of combinatorics.enumeration.

DIGITS = [1, 2]
LENGTH = 4  # the length the question asks about
BUILT = [2, 3, 5, 8]  # the counts the steps give for one digit up to four

def allowed(number):
    for place in range(len(number) - 1):
        if number[place] == 2 and number[place + 1] == 2:
            return False
    return True

def count(length):
    return len([number for number in product(DIGITS, repeat=length) if allowed(number)])

def solve(options):
    counts = [count(length) for length in range(1, LENGTH + 1)]
    if counts != BUILT:
        fail("the counts by length are %s, and the steps say %s" % (counts, BUILT))
    return match(options, counts[LENGTH - 1])
