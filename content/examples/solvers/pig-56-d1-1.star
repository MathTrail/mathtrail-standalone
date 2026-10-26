PEOPLE = 7310
LEAP = 2024  # a year with 29 February, whose days are every birthday there is

def solve(options):
    birthdays = sum([days_in_month(LEAP, month) for month in range(1, 13)])
    # The fewest people the busiest birthday can be kept to: the first number
    # that, given to every birthday, has room for the whole town.
    for most in range(1, PEOPLE + 1):
        if birthdays * most >= PEOPLE:
            return match(options, most)
    fail("the town does not fit")
