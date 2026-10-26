def solve(options):
    # Every part of the pot Al could eat, counted in 144ths so that thirds and
    # quarters of it are whole too.
    pot = 144
    parts = set()
    for al in range(0, pot + 1, 3):
        ben = al // 3
        if (pot - al) % 3 != 0:
            continue
        cy = (pot - al) // 3
        dan = pot // 4
        if al + ben + cy + dan == pot:
            common = gcd(al, pot)
            parts.add("%d/%d" % (al // common, pot // common))
    if len(parts) != 1:
        fail("%d parts fit" % len(parts))
    return match(options, list(parts)[0])
