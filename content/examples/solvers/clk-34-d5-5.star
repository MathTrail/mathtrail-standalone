def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    week = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
    days = week.index("Friday") - week.index("Monday")
    return match(options, show(at("8:00") - (2 + 3 * days)))
