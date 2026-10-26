def joined(shape):
    seen = [shape[0]]
    head = 0
    while head < len(seen):
        row, column = seen[head]
        head += 1
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            neighbour = (row + step_row, column + step_column)
            if neighbour in shape and neighbour not in seen:
                seen.append(neighbour)
    return len(seen) == len(shape)

def shape_of(cells):
    # The same shape however it is turned, flipped or moved: of its eight
    # positions, the one that sorts first, moved to the top-left corner.
    forms = []
    for turn in range(4):
        for flip in [False, True]:
            moved = []
            for row, column in cells:
                if flip:
                    column = -column
                for _ in range(turn):
                    row, column = column, -row
                moved.append((row, column))
            top = min([r for r, c in moved])
            left = min([c for r, c in moved])
            forms.append(sorted([(r - top, c - left) for r, c in moved]))
    return str(sorted(forms)[0])

def solve(options):
    # Any tile of 4 cells fits on a 4 by 4 board, so every shape is among the
    # ways of choosing 4 of its cells.
    board = [(row, column) for row in range(4) for column in range(4)]
    shapes = set()
    for cells in combinations(board, 4):
        if joined(list(cells)):
            shapes.add(shape_of(cells))
    return match(options, len(shapes))
