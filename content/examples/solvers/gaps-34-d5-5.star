def solve(options):
    tables = 8
    seats = 0
    for i in range(tables):
        seats += 2  # one place at the front of the table and one at the back
        if i == 0:
            seats += 1  # the free end of the first table
        if i == tables - 1:
            seats += 1  # and of the last
    return match(options, seats)
