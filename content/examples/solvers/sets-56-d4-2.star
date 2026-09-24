def solve(options):
    pupils = 10
    all_three = []
    for everything, swim_skate, swim_bike, skate_bike in product(range(8), repeat=4):
        swim_only = 7 - swim_skate - swim_bike - everything
        skate_only = 8 - swim_skate - skate_bike - everything
        bike_only = 9 - swim_bike - skate_bike - everything
        nothing = pupils - swim_only - skate_only - bike_only - swim_skate - swim_bike - skate_bike - everything
        if min(swim_only, skate_only, bike_only, nothing) >= 0:
            all_three.append(everything)
    return match(options, min(all_three))
