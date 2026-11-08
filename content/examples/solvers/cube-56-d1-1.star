# Fold the net square by square, the way a cube rolls over it. Each square
# keeps three directions of its face on the cube: the one pointing out of
# the cube, and the ones its right and lower sides run along. Two squares
# whose faces point out in opposite directions become opposite faces.

def flip(direction):
    return (-direction[0], -direction[1], -direction[2])

def fold(squares):
    faces = {squares[0]: ((0, 0, -1), (1, 0, 0), (0, 1, 0))}
    queue = [squares[0]]
    head = 0
    while head < len(queue):
        x, y = queue[head]
        head += 1
        out, right, down = faces[(x, y)]
        steps = [
            ((x + 1, y), (right, flip(out), down)),
            ((x - 1, y), (flip(right), out, down)),
            ((x, y + 1), (down, right, flip(out))),
            ((x, y - 1), (flip(down), right, out)),
        ]
        for square, face in steps:
            if square in squares and square not in faces:
                faces[square] = face
                queue.append(square)
    return faces

def solve(options, picture):
    squares = [tuple(square) for square in picture['cells']]
    numbers = {}
    for mark in picture.get('marks', ()):
        numbers[tuple(mark['at'])] = int(mark['label'])
    faces = fold(squares)
    outs = set([faces[square][0] for square in faces])
    if len(squares) != 6 or len(faces) != 6 or len(outs) != 6:
        fail('the squares do not fold into a cube')
    best = None
    for first in squares:
        for second in squares:
            if faces[second][0] == flip(faces[first][0]):
                total = numbers[first] + numbers[second]
                if best == None or total > best:
                    best = total
    return match(options, best)
