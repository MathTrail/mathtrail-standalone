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

def perimeter(shape):
    outside = 0
    for row, column in shape:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) not in shape:
                outside += 1
    return outside

def solve(options):
    # Any shape of 4 tiles fits on a 4 by 4 board, so every shape is among the
    # ways of choosing 4 of its squares.
    board = [(row, column) for row in range(4) for column in range(4)]
    largest = 0
    for tiles in combinations(board, 4):
        if joined(list(tiles)):
            largest = max(largest, perimeter(tiles))
    return match(options, largest)
