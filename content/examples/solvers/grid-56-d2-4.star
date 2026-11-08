def solve(options, picture):
    sheet = set(picture["cells"])
    cut = set()
    for x1, y1, x2, y2 in picture.get("lines", ()):
        if x1 == x2:
            for y in range(min(y1, y2), max(y1, y2)):
                cut.add((x1, y, "v"))
        else:
            for x in range(min(x1, x2), max(x1, x2)):
                cut.add((x, y1, "h"))
    seen = set()
    sizes = []
    for cell in sorted(sheet):
        if cell in seen:
            continue
        seen.add(cell)
        queue = [cell]
        i = 0
        while i < len(queue):
            x, y = queue[i]
            i += 1
            steps = [((x + 1, y), (x + 1, y, "v")), ((x - 1, y), (x, y, "v")), ((x, y + 1), (x, y + 1, "h")), ((x, y - 1), (x, y, "h"))]
            for nxt, edge in steps:
                if nxt in sheet and nxt not in seen and edge not in cut:
                    seen.add(nxt)
                    queue.append(nxt)
        sizes.append(len(queue))
    return match(options, len(sizes))
