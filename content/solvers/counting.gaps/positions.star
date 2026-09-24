# For posts, trees or lamps and the gaps between them: put every object at its
# place and count the places, instead of adding or taking one away by hand. A
# set keeps a place that two sides share, such as a corner, from being counted
# twice.
# From reference task gaps-34-d5-3.

SIDE = 20  # metres along each side of the square fence
APART = 4  # metres from one post to the next

def solve(options):
    posts = set()
    for x in range(0, SIDE + 1, APART):
        for y in range(0, SIDE + 1, APART):
            if x == 0 or x == SIDE or y == 0 or y == SIDE:  # on the fence, not inside the garden
                posts.add((x, y))
    return match(options, len(posts))
