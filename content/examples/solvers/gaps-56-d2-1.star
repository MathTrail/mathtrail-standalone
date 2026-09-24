WIDTH = 10  # any size will do: the cuts only have to fall inside the cake
LENGTH = 12
ALONG = [2, 5, 8]  # where the 3 cuts along the length cross the width
ACROSS = [2, 4, 7, 9]  # where the 4 cuts across cross the length

def pieces_between(size, cuts):
    # The two edges and the cuts, in order; one piece lies between each two neighbours.
    lines = sorted([0] + cuts + [size])
    return len(lines) - 1

def solve(options):
    # Every cut goes from edge to edge, so each strip along is split by every cut across.
    return match(options, pieces_between(WIDTH, ALONG) * pieces_between(LENGTH, ACROSS))
