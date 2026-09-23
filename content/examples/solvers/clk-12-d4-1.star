def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    # Tom's watch is five minutes fast, so the real time is five minutes back;
    # Kate's is five minutes slow, so hers is five minutes back from that.
    real = at("3:10") - 5
    return match(options, show(real - 5))
