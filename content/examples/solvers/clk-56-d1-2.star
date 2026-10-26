SHOWS = "4:20"
TURNS = 3  # whole turns of the minute hand

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def solve(options):
    # The clock goes on minute by minute while the minute hand turns.
    now = at(SHOWS)
    calls = 0
    for step in range(TURNS * 60):
        now += 1
        minute = now % 60
        if minute == 0:
            hour = (now // 60) % 12
            calls += hour if hour != 0 else 12
        elif minute == 30:
            calls += 1
    return match(options, calls)
