def solve(options, picture):
    top = picture["steps"]
    frog = picture["marks"][0]["step"]
    jump = picture["jumps"][0]
    size = jump["to"] - jump["from"]
    jumps = 0
    while frog < top:
        frog += size
        jumps += 1
    if frog != top:
        return match(options, "It cannot land on the top")
    return match(options, jumps)
