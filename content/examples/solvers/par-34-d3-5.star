def solve(options):
    on = {number: False for number in range(1, 10)}
    for number in range(1, 10):
        if number % 2 == 0:
            on[number] = not on[number]
    for number in range(1, 10):
        if number % 3 == 0:
            on[number] = not on[number]
    return match(options, len([number for number in range(1, 10) if on[number]]))
