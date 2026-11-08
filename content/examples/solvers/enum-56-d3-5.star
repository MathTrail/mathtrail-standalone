def routes(start, end, blocked):
    # The routes from start to end going only right or down, none through a blocked cell.
    sx, sy = start
    ex, ey = end
    ways = {}
    for y in range(sy, ey + 1):
        for x in range(sx, ex + 1):
            if (x, y) in blocked:
                ways[(x, y)] = 0
            elif (x, y) == (sx, sy):
                ways[(x, y)] = 1
            else:
                ways[(x, y)] = ways.get((x - 1, y), 0) + ways.get((x, y - 1), 0)
    return ways.get((ex, ey), 0)

def solve(options, picture):
    blocked = set(picture.get("cells", ()))
    places = {}
    for mark in picture["marks"]:
        places[mark["label"]] = mark["at"]
    return match(options, routes(places["S"], places["M"], blocked) * routes(places["M"], places["F"], blocked))
