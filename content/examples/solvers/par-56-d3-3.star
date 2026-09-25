SIZE = 3  # squares along a side of the board
CANDIDATES = [8, 10, 9, 3, 7]  # the numbers of squares the options name

def neighbours(square):
    row, column = square
    near = [(row + 1, column), (row - 1, column), (row, column + 1), (row, column - 1)]
    return [(r, c) for r, c in near if 0 <= r and r < SIZE and 0 <= c and c < SIZE]

def route_lengths():
    # Every route that steps to side neighbours, visits no square twice and
    # comes back to its first square. It visits as many squares as it makes
    # steps, which is the length of the path before the step back.
    lengths = set()
    for start in [(row, column) for row in range(SIZE) for column in range(SIZE)]:
        stack = [[start]]
        while stack:
            path = stack.pop()
            for square in neighbours(path[-1]):
                if square == start:
                    lengths.add(len(path))
                elif square not in path:
                    stack.append(path + [square])
    return lengths

def solve(options):
    lengths = route_lengths()
    fitting = [n for n in CANDIDATES if n in lengths]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
