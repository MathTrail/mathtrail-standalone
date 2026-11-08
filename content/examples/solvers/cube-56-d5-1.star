def solve(options, picture):
    path = sorted(picture["cells"])
    start = None
    for mark in picture["marks"]:
        start = mark["at"]
    told = 0
    for cell in path:
        if cell[1] == start[1] and abs(cell[0] - start[0]) % 2 == 0:
            told += 1
    return match(options, told)
