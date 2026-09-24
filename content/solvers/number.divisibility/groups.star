# For things counted out in equal groups, one group after another: try every
# way of writing a thing's number as full groups and then a place in the next
# group, counting places from 1, and read off the group, the place or the
# number of full groups.
# From reference task div-56-d3-1.

NUMBER = 97  # the thing the question names: here, seat 97
SIZE = 18  # how many things each group holds: here, seats in a row

def solve(options):
    # Places count from 1, so the last thing of a group is its SIZE-th place,
    # never place 0 of the next group.
    splits = [(full, place) for full in range(0, NUMBER + 1) for place in range(1, SIZE + 1)
              if full * SIZE + place == NUMBER]
    if len(splits) != 1:
        fail("%d ways of placing %d" % (len(splits), NUMBER))
    full, place = splits[0]
    return match(options, full + 1)  # the group it is in; place for where it is within that group
