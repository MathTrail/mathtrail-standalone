def cell_squares(cells, width, height):
    # The squares of every size made only of filled cells.
    found = []
    for size in range(1, min(width, height) + 1):
        for x in range(width - size + 1):
            for y in range(height - size + 1):
                if all([(x + i, y + j) in cells for i in range(size) for j in range(size)]):
                    found.append((x, y, size))
    return found

def solve(options, picture):
    cells = set(picture["cells"])
    x = None
    for mark in picture["marks"]:
        if mark["label"] == "X":
            x = mark["at"]
    count = 0
    for left, top, size in cell_squares(cells, picture["width"], picture["height"]):
        if left <= x[0] and x[0] < left + size and top <= x[1] and x[1] < top + size:
            count += 1
    return match(options, count)
