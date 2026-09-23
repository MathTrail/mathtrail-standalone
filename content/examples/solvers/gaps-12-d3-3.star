def solve(options):
    line = ["ahead"] * 3 + ["Kate"] + ["behind"] * 3
    if line.index("Kate") + 1 != 4:
        fail("Kate does not stand fourth in this line")
    return match(options, len(line))
