FIRST = 3  # the date of the first meeting in May
EVERY = 4  # the days from one meeting to the next
MEETING = 5  # the meeting the question asks about

def solve(options):
    meetings = [day for day in range(FIRST, days_in_month(2026, 5) + 1) if (day - FIRST) % EVERY == 0]
    if len(meetings) < MEETING:
        fail("May has only %d meetings" % len(meetings))
    return match(options, "%d May" % meetings[MEETING - 1])
