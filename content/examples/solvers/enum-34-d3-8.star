def solve(options, picture):
    width = picture["width"]
    height = picture["height"]
    blocked = set()
    for x1, y1, x2, y2 in picture.get("lines", ()):
        if x1 == x2:
            for y in range(min(y1, y2), max(y1, y2)):
                blocked.add((x1, y, "v"))
        elif y1 == y2:
            for x in range(min(x1, x2), max(x1, x2)):
                blocked.add((x, y1, "h"))
    places = {}
    for mark in picture["marks"]:
        places[mark["label"]] = mark["at"]
    start = places["S"]
    dist = {start: 0}
    queue = [start]
    i = 0
    while i < len(queue):
        x, y = queue[i]
        i += 1
        for dx, dy in [(1, 0), (-1, 0), (0, 1), (0, -1)]:
            nx = x + dx
            ny = y + dy
            if nx < 0 or ny < 0 or nx >= width or ny >= height:
                continue
            if dx == 1 and (x + 1, y, "v") in blocked:
                continue
            if dx == -1 and (x, y, "v") in blocked:
                continue
            if dy == 1 and (x, y + 1, "h") in blocked:
                continue
            if dy == -1 and (x, y, "h") in blocked:
                continue
            if (nx, ny) not in dist:
                dist[(nx, ny)] = dist[(x, y)] + 1
                queue.append((nx, ny))
    if places["E"] not in dist:
        return match(options, "There is no way")
    return match(options, dist[places["E"]])
