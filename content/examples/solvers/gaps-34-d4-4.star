def flights(start, end):
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    per_flight = 36 // flights(1, 3)
    return match(options, flights(3, 7) * per_flight)
