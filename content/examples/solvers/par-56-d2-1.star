START = 1
CANDIDATES = [51, 64, 40, 100, 52]  # the numbers the options name

def solve(options):
    # Every number the steps can make, up to the largest option: both steps
    # only make a number bigger, so nothing past it can lead back down.
    reached = set([START])
    queue = [START]
    head = 0
    while head < len(queue):
        number = queue[head]
        head += 1
        for made in [number + 4, number * 3]:
            if made <= max(CANDIDATES) and made not in reached:
                reached.add(made)
                queue.append(made)
    fitting = [n for n in CANDIDATES if n in reached]
    if len(fitting) != 1:
        fail("%d of the numbers can be reached" % len(fitting))
    return match(options, fitting[0])
