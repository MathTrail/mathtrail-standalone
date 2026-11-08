def solve(options, picture):
    figure = set(picture["cells"])
    x1, y1, x2, y2 = picture["dashed"][0]
    need = 0
    for x, y in figure:
        if x1 == x2:
            image = (2 * x1 - 1 - x, y)
        else:
            image = (x, 2 * y1 - 1 - y)
        if image not in figure:
            need += 1
    return match(options, need)
