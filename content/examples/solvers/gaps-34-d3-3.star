def solve(options):
    line = ["ahead"] * 6 + ["Masha"] + ["between"] * 3 + ["Dasha"] + ["behind"] * 4
    if line.index("Masha") + 1 != 7:
        fail("Masha does not stand seventh from the front")
    if len(line) - line.index("Dasha") != 5:
        fail("Dasha does not stand fifth from the back")
    return match(options, len(line))
