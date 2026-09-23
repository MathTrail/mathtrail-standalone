def solve(options):
    circle = ["Ann", "Ben", "Kim", "Dan", "Eva", "Fay"]
    index = 0
    while len(circle) > 1:
        index = (index + 1) % len(circle)  # one child is skipped, the next steps out
        circle.pop(index)
        index = index % len(circle)
    return match(options, circle[0])
