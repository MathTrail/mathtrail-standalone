DAYS = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
BEIJING = ("Saturday", "1:00")
LONDON_AND_MOSCOW = ("12:00", "15:00")  # one moment on the clocks of London and of Moscow
MOSCOW_AND_BEIJING = ("15:00", "20:00")  # one moment on the clocks of Moscow and of Beijing

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the options write it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    moscow_ahead_of_london = at(LONDON_AND_MOSCOW[1]) - at(LONDON_AND_MOSCOW[0])
    beijing_ahead_of_moscow = at(MOSCOW_AND_BEIJING[1]) - at(MOSCOW_AND_BEIJING[0])
    beijing = DAYS.index(BEIJING[0]) * 24 * 60 + at(BEIJING[1])  # minutes from the start of the week
    london = beijing - beijing_ahead_of_moscow - moscow_ahead_of_london
    return match(options, show(london) + " on " + DAYS[(london // (24 * 60)) % 7])
