def possible(friends):
    # Every handshake uses two hands, so three hands a friend must pair up.
    if (3 * friends) % 2 == 1:
        return False
    # For an even number, a ring in which each friend also shakes the hand of
    # the one opposite gives everybody exactly three.
    shakes = set()
    for i in range(friends):
        for j in [(i + 1) % friends, (i + friends // 2) % friends]:
            shakes.add((min([i, j]), max([i, j])))
    return all([len([shake for shake in shakes if i in shake]) == 3 for i in range(friends)])

def solve(options):
    fitting = [n for n in [5, 7, 8, 9, 11] if possible(n)]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
