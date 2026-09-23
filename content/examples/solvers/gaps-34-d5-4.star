def flights(start, end):
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    per_flight = 60 // flights(1, 5)
    return match(options, flights(3, 9) * per_flight)
