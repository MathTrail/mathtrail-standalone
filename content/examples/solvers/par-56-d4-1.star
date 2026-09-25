START = (0, 1, 2)  # frogs P, Q and R
CANDIDATES = [(0, 2, 4), (0, 4, 5), (1, 2, 5), (0, 1, 3), (1, 3, 5)]  # as the options list them
WIDE = 12  # the search keeps every frog between -WIDE and WIDE

def places_reached(loose):
    # Every way the frogs can sit, breadth first. The task does not say whether
    # a frog may pass over both others in one jump or land on a frog: a loose
    # search lets it do both, a strict one neither. Kept between -WIDE and WIDE,
    # the search shows that the answer can be reached; that the other options
    # cannot, even by jumps further out, is what each frog keeping the parity
    # of its point says.
    seen = set([START])
    queue = [START]
    head = 0
    while head < len(queue):
        frogs = queue[head]
        head += 1
        for jumper in range(3):
            for over in range(3):
                if jumper == over:
                    continue
                land = 2 * frogs[over] - frogs[jumper]
                third = frogs[3 - jumper - over]
                low, high = min([frogs[jumper], land]), max([frogs[jumper], land])
                if abs(land) > WIDE or not loose and (third == land or low < third and third < high):
                    continue
                moved = tuple([land if frog == jumper else place for frog, place in enumerate(frogs)])
                if moved not in seen:
                    seen.add(moved)
                    queue.append(moved)
    return set([tuple(sorted(frogs)) for frogs in queue])

def solve(options):
    # The answer must not depend on how the jumps are read.
    strict, loose = places_reached(False), places_reached(True)
    strict = [places for places in CANDIDATES if places in strict]
    loose = [places for places in CANDIDATES if places in loose]
    if strict != loose:
        fail("the strict and the loose search disagree: %s against %s" % (strict, loose))
    if len(strict) != 1:
        fail("%d of the options can be reached" % len(strict))
    return match(options, "%d, %d and %d" % strict[0])
