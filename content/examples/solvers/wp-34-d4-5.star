def scale(heavy_bag):
    # One coin from bag 1, two from bag 2, and so on; the heavy bag's coins
    # weigh 11 g and every other coin 10 g.
    return sum([bag * (11 if bag == heavy_bag else 10) for bag in range(1, 11)])

def solve(options):
    fitting = ["bag %d" % bag for bag in range(1, 11) if scale(bag) == 553]
    if len(fitting) != 1:
        fail("%d of the bags fit what the scale shows" % len(fitting))
    return match(options, fitting[0])
