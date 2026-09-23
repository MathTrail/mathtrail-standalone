def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def strikes(start, end, half=False, quarters=False):
    struck = 0
    for now in range(start, end + 1):
        minute = now % 60
        if minute == 0:
            hour = (now // 60) % 12
            if hour == 0:
                hour = 12
            struck += hour
        elif half and minute == 30:
            struck += 1
        elif quarters and (minute == 15 or minute == 30 or minute == 45):
            struck += 1
    return struck

def solve(options):
    start = at("12:15")
    return match(options, strikes(start, start + 12 * 60 - 1, half=True))
