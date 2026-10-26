YEAR = 2026

def even_digits(number):
    return all([int(digit) % 2 == 0 for digit in str(number).elems()])

def solve(options):
    # Every date of the year, its day, its month and its year written out.
    if not even_digits(YEAR):
        fail("%d has an odd digit" % YEAR)
    even = 0
    for month in range(1, 13):
        for day in range(1, days_in_month(YEAR, month) + 1):
            if even_digits(day) and even_digits(month):
                even += 1
    return match(options, even)
