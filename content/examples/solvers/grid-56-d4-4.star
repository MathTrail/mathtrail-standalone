# Number the squares by their steps from the peg, layer by layer, going
# round the shed, and count those the number on the peg's square allows.
def solve(options, picture):
    width = picture['width']
    height = picture['height']
    shed = [(cell[0], cell[1]) for cell in picture['cells']]
    marks = picture['marks']
    if len(marks) != 1:
        fail('the picture marks the peg square alone, with the steps of the leash')
    peg = (marks[0]['at'][0], marks[0]['at'][1])
    steps = int(marks[0]['label'])
    reached = [peg]
    edge = [peg]
    for _ in range(steps):
        following = []
        for x, y in edge:
            for dx, dy in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
                cell = (x + dx, y + dy)
                inside = cell[0] >= 0 and cell[0] < width and cell[1] >= 0 and cell[1] < height
                if inside and cell not in shed and cell not in reached:
                    reached.append(cell)
                    following.append(cell)
        edge = following
    return match(options, len(reached))
