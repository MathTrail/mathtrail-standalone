YEAR = 2015  # any year would do: every month has at least 28 days, so every day of the week
WANTED = 2  # pupils born in the same month on the same day of the week

def solve(options):
    # The kinds of pupil are the pairs of a month and a day of the week that a
    # birthday can fall on.
    kinds = set()
    for month in range(1, 13):
        for day in range(1, days_in_month(YEAR, month) + 1):
            kinds.add((month, weekday(YEAR, month, day)))
    # Every size a school can have with fewer than WANTED pupils of each kind,
    # built up one kind at a time; the answer is the first size not among them.
    failed = set([0])
    for _ in kinds:
        failed = set([size + count for size in failed for count in range(WANTED)])
    pupils = 0
    while pupils in failed:
        pupils += 1
    return match(options, pupils)
