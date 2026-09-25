TO_TOWN = ["first", "second", "third"]  # roads between the village and the town
TO_CITY = ["fourth", "fifth"]  # roads between the town and the city

def allowed(trip):
    # No road is used twice.
    return len(set(trip)) == len(trip)

def solve(options):
    trips = product(TO_TOWN, TO_CITY, TO_CITY, TO_TOWN)  # there, then back
    return match(options, len([trip for trip in trips if allowed(trip)]))
