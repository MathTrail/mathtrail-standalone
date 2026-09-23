def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def solve(options):
    departures = range(at("6:10"), at("10:00"), 25)
    within = [time for time in departures if at("7:00") <= time and time <= at("9:00")]
    return match(options, len(within))
