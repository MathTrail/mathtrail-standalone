def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    now = at("8:30")
    for lesson in range(1, 4):
        now += 40
        if lesson < 3:
            now += 10
    return match(options, show(now))
