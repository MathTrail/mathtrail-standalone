def triangles(lines):
    # The triangles whose sides lie on the lines. Every line is cut into steps from
    # corner to corner of the paper, so that lines laid end to end along one
    # straight line join their corners as one line does; two corners are joined
    # when every step between them is drawn.
    steps = set()
    points = set()
    for x1, y1, x2, y2 in lines:
        dx, dy = x2 - x1, y2 - y1
        n = gcd(abs(dx), abs(dy))
        run = [(x1 + k * dx // n, y1 + k * dy // n) for k in range(n + 1)]
        for k in range(n):
            steps.add((run[k], run[k + 1]))
            steps.add((run[k + 1], run[k]))
        for p in run:
            points.add(p)
    def joined(a, b):
        dx, dy = b[0] - a[0], b[1] - a[1]
        n = gcd(abs(dx), abs(dy))
        sx, sy = dx // n, dy // n
        for k in range(n):
            if ((a[0] + k * sx, a[1] + k * sy), (a[0] + (k + 1) * sx, a[1] + (k + 1) * sy)) not in steps:
                return False
        return True
    points = sorted(points)
    count = 0
    for i in range(len(points)):
        for j in range(i + 1, len(points)):
            if not joined(points[i], points[j]):
                continue
            for k in range(j + 1, len(points)):
                a, b, c = points[i], points[j], points[k]
                if (b[0] - a[0]) * (c[1] - a[1]) != (b[1] - a[1]) * (c[0] - a[0]) and joined(a, c) and joined(b, c):
                    count += 1
    return count

def solve(options, picture):
    return match(options, triangles(picture["lines"]))
