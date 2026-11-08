# A cross of six squares: a column of 4 numbered 1 to 4 from the top, and 5
# and 6 on either side of square 2. Roll a cube over the squares from the
# first, so that each square lands on the face the roll brings to the bottom,
# and find the square that lands opposite square 2.
# Written from the template fold-net of geometry.cubes_paper.

NET = {(1, 0): 1, (1, 1): 2, (1, 2): 3, (1, 3): 4, (0, 1): 5, (2, 1): 6}  # (column, row): number
ASKED = 2  # the number whose opposite face is asked for

OPPOSITE = {0: 1, 1: 0, 2: 3, 3: 2, 4: 5, 5: 4}

def fold(net):
    cells = sorted(net)
    faces = {cells[0]: {"bottom": 0, "top": 1, "north": 2, "south": 3, "east": 4, "west": 5}}
    queue = [cells[0]]
    i = 0
    while i < len(queue):
        x, y = queue[i]
        i += 1
        o = faces[(x, y)]
        rolls = [((x + 1, y), {"bottom": o["east"], "east": o["top"], "top": o["west"], "west": o["bottom"], "north": o["north"], "south": o["south"]}),
                 ((x - 1, y), {"bottom": o["west"], "west": o["top"], "top": o["east"], "east": o["bottom"], "north": o["north"], "south": o["south"]}),
                 ((x, y + 1), {"bottom": o["south"], "south": o["top"], "top": o["north"], "north": o["bottom"], "east": o["east"], "west": o["west"]}),
                 ((x, y - 1), {"bottom": o["north"], "north": o["top"], "top": o["south"], "south": o["bottom"], "east": o["east"], "west": o["west"]})]
        for cell, turned in rolls:
            if cell in net and cell not in faces:
                faces[cell] = turned
                queue.append(cell)
    return faces

def solve(options):
    faces = fold(NET)
    asked = [cell for cell in NET if NET[cell] == ASKED][0]
    for cell in NET:
        if faces[cell]["bottom"] == OPPOSITE[faces[asked]["bottom"]]:
            return match(options, NET[cell])
    return []
