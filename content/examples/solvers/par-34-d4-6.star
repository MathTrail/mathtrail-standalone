def solve(options, picture):
    sides = picture["sides"]
    joint = picture.get("joint", sides // 2)
    tiles = picture["tiles"]
    found = []
    shown = [int(t) for t in tiles[0] if t != "" and t != "?"]
    for rest in product(range(1, 10), repeat=sides - len(shown)):
        base = shown + list(rest)
        if len(set(base)) != sides:
            continue
        shift = 0
        fits = True
        asked = None
        for tile in tiles:
            for k in range(sides):
                value = base[(k + shift) % sides]
                if tile[k] == "?":
                    asked = value
                elif tile[k] != "" and int(tile[k]) != value:
                    fits = False
            shift = base.index(base[(joint + shift) % sides])
        if fits and asked not in found:
            found.append(asked)
    if len(found) != 1:
        return match(options, "It cannot be told")
    return match(options, found[0])
