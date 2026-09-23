def solve(options):
    grandpa = 60
    dad = 35
    while dad > 0:
        grandpa -= 1
        dad -= 1
    return match(options, grandpa)
