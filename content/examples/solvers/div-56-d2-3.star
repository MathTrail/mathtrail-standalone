def solve(options):
    stations = 20
    # Move round the loop one station at a time: after the last comes the first.
    station = 1
    for _ in range(47):
        station = station + 1 if station < stations else 1
    return match(options, station)
