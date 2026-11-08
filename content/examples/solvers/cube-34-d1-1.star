def fold(net):
    # The face of the cube each cell of the net lands on, rolling the cube from cell to cell.
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

OPPOSITE = {0: 1, 1: 0, 2: 3, 3: 2, 4: 5, 5: 4}

def opposite_cell(faces, cell):
    # The cell of the net that lands opposite this one, or None.
    for other in faces:
        if faces[other]["bottom"] == OPPOSITE[faces[cell]["bottom"]]:
            return other
    return None

def is_cube(net):
    # Whether the cells fold into a cube, each face covered once.
    faces = fold(net)
    if len(net) != 6 or len(faces) != 6:
        return False
    return len(set([faces[cell]["bottom"] for cell in faces])) == 6

def solve(options, picture):
    net = set(picture["cells"])
    faces = fold(net)
    present = [faces[cell]["bottom"] for cell in faces]
    lid = [face for face in range(6) if face not in present]
    if len(net) != 5 or len(faces) != 5 or len(lid) != 1:
        return []
    for mark in picture["marks"]:
        if faces[mark["at"]]["bottom"] == OPPOSITE[lid[0]]:
            return match(options, int(mark["label"]))
    return []
