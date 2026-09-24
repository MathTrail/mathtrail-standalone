SIDE = 12  # metres along each side of the square garden
APART = 3  # metres from one post to the next
MIDDLE = SIDE // 2  # where the two inner fences run

def solve(options):
    posts = set()
    for x in range(0, SIDE + 1, APART):
        for y in range(0, SIDE + 1, APART):
            outer = x == 0 or x == SIDE or y == 0 or y == SIDE
            inner = x == MIDDLE or y == MIDDLE
            if outer or inner:  # posts stand on fences, and a set keeps a shared one once
                posts.add((x, y))
    return match(options, len(posts))
