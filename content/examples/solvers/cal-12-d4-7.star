STAYS = [3, 2]  # the nights in the hotel, then at grandma's straight after

def solve(options):
    # A stay runs from the day it starts to the day after its last night, and
    # the day the family moves ends one stay as it starts the next.
    days = []
    start = 1
    for nights in STAYS:
        for day in range(start, start + nights + 1):
            if day not in days:
                days.append(day)
        start += nights
    return match(options, len(days))
