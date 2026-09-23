def solve(options):
    shown = 0
    for now in range(24 * 60):
        if now // 60 == now % 60:
            shown += 1
    return match(options, shown)
