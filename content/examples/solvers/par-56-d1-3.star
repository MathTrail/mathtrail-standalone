CANDIDATES = [6, 15, 21, 35, 10]  # the numbers the options name

def solve(options):
    # Every pair of different whole numbers whose result could be one of the
    # options: the result is never smaller than the larger number.
    made = set()
    for big in range(2, max(CANDIDATES) + 1):
        for small in range(1, big):
            made.add((big - small) * big * small)
    fitting = [n for n in CANDIDATES if n in made]
    if len(fitting) != 1:
        fail("%d of the numbers can be made" % len(fitting))
    return match(options, fitting[0])
