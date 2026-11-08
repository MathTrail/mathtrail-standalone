def solve(options, picture):
    sides = picture["sides"]
    joint = picture.get("joint", sides // 2)
    tiles = picture["tiles"]
    base = list(tiles[0])
    if "" in base or "?" in base:
        return []
    found = []
    shift = 0
    for tile in tiles:
        for k in range(sides):
            value = base[(k + shift) % sides]
            if tile[k] == "?":
                if value not in found:
                    found.append(value)
            elif tile[k] != "" and tile[k] != value:
                return []
        shift = base.index(base[(joint + shift) % sides])
    if picture["layout"] == "ring" and shift != 0:
        return []
    if len(found) != 1:
        return match(options, "It cannot be told")
    return match(options, int(found[0]))
