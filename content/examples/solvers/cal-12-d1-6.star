MONTHS = ["January", "February", "March", "April", "May", "June", "July", "August",
          "September", "October", "November", "December"]

def later(month, months):
    # One month at a time, past December into a new year.
    index = MONTHS.index(month)
    for _ in range(months):
        index = (index + 1) % 12
    return MONTHS[index]

def solve(options):
    return match(options, later("November", 2))
