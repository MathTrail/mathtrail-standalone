# A hundred twos multiplied together: the last digits of the products go round
# a pattern 4 long, so write 100 as full rounds of the pattern and then a place
# in the next round, counting places from 1, and read the digit at that place.
# The pattern is checked to repeat, and the digit against the hundred twos
# multiplied out in full.
# Written from the template groups of number.divisibility.

TWOS = 100
SIZE = 4  # the length of the pattern of last digits

def solve(options):
    digits = []  # the last digit after 1, 2, 3 ... twos
    last = 1
    for _ in range(2 * SIZE):
        last = last * 2 % 10
        digits.append(last)
    if digits[:SIZE] != digits[SIZE:]:
        fail("the last digits do not repeat every %d twos" % SIZE)
    splits = [(full, place) for full in range(0, TWOS + 1) for place in range(1, SIZE + 1)
              if full * SIZE + place == TWOS]
    if len(splits) != 1:
        fail("%d ways of placing the %dth two" % (len(splits), TWOS))
    full, place = splits[0]
    whole = 1
    for _ in range(TWOS):
        whole *= 2
    if whole % 10 != digits[place - 1]:
        fail("the pattern gives %d, and the hundred twos end in %d" % (digits[place - 1], whole % 10))
    return match(options, digits[place - 1])
