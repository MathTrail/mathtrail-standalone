# 150 pupils go on a trip in buses of 40 seats: write the last pupil's number
# as full buses and then a seat in the next bus, counting seats from 1, and
# read off the bus that pupil rides in, which is how many buses are needed.
# Written from the template groups of number.divisibility.

NUMBER = 150  # the last pupil
SIZE = 40  # the seats in one bus

def solve(options):
    # Seats count from 1, so the last pupil of a full bus sits in its SIZE-th
    # seat, never in seat 0 of the next bus.
    splits = [(full, place) for full in range(0, NUMBER + 1) for place in range(1, SIZE + 1)
              if full * SIZE + place == NUMBER]
    if len(splits) != 1:
        fail("%d ways of seating pupil %d" % (len(splits), NUMBER))
    full, place = splits[0]
    if (full, place) != (3, 30):
        fail("the steps say 3 full buses and 30 pupils left over")
    return match(options, full + 1)
