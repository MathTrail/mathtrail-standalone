def flights(start, end):
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    per_flight = 30 // flights(1, 4)
    return match(options, flights(1, 10) * per_flight)
