def solve(options, picture):
    rows = picture["rows"]
    names = []
    for row in rows:
        for shape in row["shapes"]:
            if shape not in names:
                names.append(shape)
    found = []
    for values in product(range(0, 21), repeat=len(names)):
        value = dict(zip(names, values))
        fits = True
        asked = None
        for row in rows:
            total = 0
            for shape in row["shapes"]:
                total += value[shape]
            if row["total"] == "?":
                asked = total
            elif total != int(row["total"]):
                fits = False
                break
        if fits and asked not in found:
            found.append(asked)
    if len(found) != 1:
        return []
    return match(options, found[0])
