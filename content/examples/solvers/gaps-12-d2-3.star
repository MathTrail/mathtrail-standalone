def line_length(place_from_front, place_from_back):
    for people in range(1, 500):
        if people - place_from_front + 1 == place_from_back:
            return people
    fail("no line gives both places to one person")

def solve(options):
    return match(options, line_length(3, 3))
