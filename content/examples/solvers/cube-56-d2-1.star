def solve(options, picture):
    path = set(picture["cells"])
    start = None
    end = None
    top = None
    for mark in picture["marks"]:
        if mark["label"] == "S":
            start = mark["at"]
        else:
            end = mark["at"]
            top = int(mark["label"])
    if start[0] != end[0] and start[1] != end[1]:
        return match(options, "It cannot be told")
    rolls = abs(end[0] - start[0]) + abs(end[1] - start[1])
    for x in range(min(start[0], end[0]), max(start[0], end[0]) + 1):
        for y in range(min(start[1], end[1]), max(start[1], end[1]) + 1):
            if (x, y) not in path:
                return []
    if rolls % 4 == 0:
        return match(options, top)
    if rolls % 4 == 2:
        return match(options, 7 - top)
    return match(options, "It cannot be told")
