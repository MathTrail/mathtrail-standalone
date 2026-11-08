def forms_of(piece):
    forms = []
    cells = list(piece)
    for turn in range(4):
        left = min([c[0] for c in cells])
        top = min([c[1] for c in cells])
        form = sorted([(c[0] - left, c[1] - top) for c in cells])
        if form not in forms:
            forms.append(form)
        cells = [(-c[1], c[0]) for c in cells]
    return forms

def coverings(shape, forms, limit):
    order = sorted(shape)
    found = 0
    stack = [()]
    while stack and found < limit:
        covered = stack.pop()
        held = set(covered)
        first = None
        for cell in order:
            if cell not in held:
                first = cell
                break
        if first == None:
            found += 1
            continue
        for form in forms:
            dx = first[0] - form[0][0]
            dy = first[1] - form[0][1]
            placed = [(c[0] + dx, c[1] + dy) for c in form]
            fits = True
            for cell in placed:
                if cell not in shape or cell in held:
                    fits = False
                    break
            if fits:
                stack.append(covered + tuple(placed))
    return found

def solve(options, picture):
    shape = set(picture["cells"])
    piece = []
    for mark in picture["marks"]:
        if mark["label"] == "1":
            piece.append(mark["at"])
    forms = forms_of(piece)
    if len(shape) % len(piece) != 0 or coverings(shape, forms, 1) == 0:
        return match(options, "It cannot be done")
    return match(options, len(shape) // len(piece))
