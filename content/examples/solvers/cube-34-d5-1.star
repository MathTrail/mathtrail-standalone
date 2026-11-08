def solve(options, picture):
    path = set(picture["cells"])
    start = None
    end = None
    top = None
    for mark in picture["marks"]:
        if mark["label"] == "E":
            end = mark["at"]
        else:
            start = mark["at"]
            top = int(mark["label"])
    rolls = {start: 0}
    queue = [start]
    i = 0
    while i < len(queue):
        x, y = queue[i]
        i += 1
        for nxt in [(x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)]:
            if nxt in path and nxt not in rolls:
                rolls[nxt] = rolls[(x, y)] + 1
                queue.append(nxt)
    if end not in rolls or (start[0] != end[0] and start[1] != end[1]):
        return match(options, "It cannot be told")
    k = rolls[end] % 4
    if k == 0:
        return match(options, top)
    if k == 2:
        return match(options, 7 - top)
    return match(options, "It cannot be told")
