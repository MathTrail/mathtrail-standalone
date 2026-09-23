def pad(number):
    return str(number) if number >= 10 else "0" + str(number)

def solve(options):
    shown = 0
    for now in range(24 * 60):
        face = pad(now // 60) + pad(now % 60)
        if len(set(face.elems())) == 1:
            shown += 1
    return match(options, shown)
