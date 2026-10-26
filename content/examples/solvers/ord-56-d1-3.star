SHAPES = ["circle", "square", "star", "heart"]
COLOURS = ["Red", "Blue", "Green", "Yellow"]

def solve(options):
    # Every order of the shapes and every way of colouring them.
    circles = set()
    for order in permutations(SHAPES):
        at = {shape: i for i, shape in enumerate(order)}
        if at["circle"] != 0 or at["heart"] != 3:
            continue
        for painted in permutations(COLOURS):
            colour = {shape: painted[i] for i, shape in enumerate(order)}
            red = [i for i in range(4) if painted[i] == "Red"][0]
            green = [i for i in range(4) if painted[i] == "Green"][0]
            if red == at["circle"] + 1 and colour["star"] == "Yellow" and abs(green - at["star"]) == 1:
                circles.add(colour["circle"])
    if not circles:
        fail("no row fits")
    if len(circles) != 1:
        return match(options, "It cannot be told")
    return match(options, list(circles)[0])
