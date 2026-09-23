def solve(options):
    places = set()
    for order in permutations(["Tom", "Sam", "c1", "c2", "c3"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Sam"] - place["Tom"] == 3:
            places.add(place["Tom"])
    return match(options, len(places))
