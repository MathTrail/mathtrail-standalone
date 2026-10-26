PARALLELS = 9  # circles round the globe, evenly spaced from pole to pole
MERIDIANS = 12  # lines from pole to pole, evenly spaced round it
DOWN = 3  # degrees from the North Pole that a cell of the map spans
ROUND = 5  # degrees round the globe that a cell spans

def solve(options):
    # The surface as a map of small cells, by how far down from the North Pole
    # and how far round it they lie. Two cells side by side are in one part
    # unless a parallel or a meridian runs between them, and the parts are
    # found by walking from cell to cell; the meridians meet at each pole, so
    # no walk crosses a pole.
    parallels = set([180 * k // (PARALLELS + 1) for k in range(1, PARALLELS + 1)])
    meridians = set([360 * k // MERIDIANS for k in range(MERIDIANS)])
    if any([p % DOWN != 0 for p in parallels]) or any([m % ROUND != 0 for m in meridians]):
        fail("a line runs through cells of the map rather than between them")
    rows, cols = 180 // DOWN, 360 // ROUND
    part = {}
    parts = 0
    for row in range(rows):
        for col in range(cols):
            if (row, col) in part:
                continue
            parts += 1
            part[(row, col)] = parts
            queue = [(row, col)]
            head = 0
            while head < len(queue):
                y, x = queue[head]
                head += 1
                near = []
                if y > 0 and y * DOWN not in parallels:
                    near.append((y - 1, x))
                if y < rows - 1 and (y + 1) * DOWN not in parallels:
                    near.append((y + 1, x))
                if x * ROUND not in meridians:
                    near.append((y, (x - 1) % cols))
                if ((x + 1) % cols) * ROUND not in meridians:
                    near.append((y, (x + 1) % cols))
                for cell in near:
                    if cell not in part:
                        part[cell] = parts
                        queue.append(cell)
    return match(options, parts)
