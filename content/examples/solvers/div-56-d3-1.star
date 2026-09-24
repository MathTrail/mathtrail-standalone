def solve(options):
    seat = 97
    per_row = 18
    # Every way of writing the seat as full rows and then a place in the next
    # row, counting places from 1: seat 90 is the 18th of row 5, not the 0th of row 6.
    splits = [(rows, place) for rows in range(0, seat + 1) for place in range(1, per_row + 1)
              if rows * per_row + place == seat]
    if len(splits) != 1:
        fail("%d ways of placing the seat" % len(splits))
    full_rows, _ = splits[0]
    return match(options, full_rows + 1)
