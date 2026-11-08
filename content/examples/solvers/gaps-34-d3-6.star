def solve(options, picture):
    top = picture["steps"]
    start = {}
    for mark in picture["marks"]:
        start[mark["label"]] = mark["step"]
    up = 0
    down = 0
    for jump in picture["jumps"]:
        if jump["from"] == start["A"]:
            up = jump["to"] - jump["from"]
        if jump["from"] == start["B"]:
            down = jump["from"] - jump["to"]
    a = start["A"]
    b = start["B"]
    moves = 0
    while a < b and moves <= top:
        a += up
        b -= down
        moves += 1
    if a != b:
        return match(options, "They never meet")
    return match(options, a)
