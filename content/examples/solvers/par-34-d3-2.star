def solve(options):
    sides = ["girl", "boy"]
    lines = [[sides[(place + start) % 2] for place in range(25)] for start in [0, 1]]
    more_girls = []
    for line in lines:
        girls = len([child for child in line if child == "girl"])
        if girls > len(line) - girls:
            more_girls.append(girls)
    if len(more_girls) != 1:
        fail("%d of the lines have more girls than boys" % len(more_girls))
    return match(options, more_girls[0])
