def solve(options):
    bucket = 10
    jug = 0
    capacity = 7
    poured = min([bucket, capacity - jug])
    return match(options, bucket - poured)
