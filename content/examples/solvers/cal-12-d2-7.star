WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
DAYS_OFF = ["Saturday", "Sunday"]
WEEKS = 3

def solve(options):
    days = [WEEK[n % 7] for n in range(7 * WEEKS)]
    return match(options, len([day for day in days if day in DAYS_OFF]))
