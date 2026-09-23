def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def solve(options):
    ends = []
    now = at("8:00")
    while now < at("13:00"):
        now += 45
        ends.append(now)
        now += 10
    return match(options, len([end for end in ends if end < at("12:00")]))
