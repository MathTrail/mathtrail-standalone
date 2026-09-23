def flights(start, end):
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    return match(options, flights(1, 9) * 18)
