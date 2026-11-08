def layers_after(picture, folds):
    # How many layers lie on each cell once the sheet is folded along the folds in turn.
    layers = {}
    for cell in picture["cells"]:
        layers[cell] = 1
    for x1, y1, x2, y2 in folds:
        moved = {}
        for cell in layers:
            x, y = cell
            if x1 == x2 and x < x1:
                cell = (2 * x1 - 1 - x, y)
            elif y1 == y2 and y < y1:
                cell = (x, 2 * y1 - 1 - y)
            moved[cell] = moved.get(cell, 0) + layers[(x, y)]
        layers = moved
    return layers

def solve(options, picture):
    order = {}
    for point in picture.get("points", ()):
        order[point["at"]] = int(point["label"])
    folds = sorted([(order.get((x1, y1), order.get((x2, y2), 0)), (x1, y1, x2, y2)) for x1, y1, x2, y2 in picture["dashed"]])
    layers = layers_after(picture, [fold for _, fold in folds])
    hole = picture["marks"][0]["at"]
    return match(options, layers.get(hole, 0))
