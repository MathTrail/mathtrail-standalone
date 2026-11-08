def solve(options, picture):
    lengths = []
    for x1, y1, x2, y2 in picture["lines"]:
        lengths.append(max(abs(x2 - x1), abs(y2 - y1)))
    total = sum(lengths)
    if total % 4 != 0:
        return match(options, "It is impossible")
    side = total // 4
    for places in product(range(4), repeat=len(lengths)):
        sums = [0, 0, 0, 0]
        for i in range(len(lengths)):
            sums[places[i]] += lengths[i]
        if sums == [side, side, side, side]:
            return match(options, side)
    return match(options, "It is impossible")
