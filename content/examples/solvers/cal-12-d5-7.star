MONTHS = ["January", "February", "March", "April", "May", "June", "July", "August",
          "September", "October", "November", "December"]
TODAY = (2026, 5, 25)  # any year: May has 31 days and June 30 in every one
DAYS = 40

def solve(options):
    year, month, day = add_days(TODAY, DAYS)
    return match(options, "%d %s" % (day, MONTHS[month - 1]))
