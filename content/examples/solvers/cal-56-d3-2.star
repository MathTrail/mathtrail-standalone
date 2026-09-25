FIRST = 1896
LAST = 1912  # both counted

def leap(year):
    # The rule as the task states it.
    return year % 4 == 0 and (year % 100 != 0 or year % 400 == 0)

def solve(options):
    years = range(FIRST, LAST + 1)
    if [year for year in years if leap(year) != is_leap(year)]:
        fail("the rule in the task disagrees with the calendar")
    return match(options, len([year for year in years if leap(year)]))
