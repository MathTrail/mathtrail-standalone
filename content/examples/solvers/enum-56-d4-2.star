FRIENDS = ["Nora", "Owen", "Pia", "Raj", "Sam"]  # on the bench, seats 1 to 5

def allowed(row):
    seat = {name: i + 1 for i, name in enumerate(row)}
    # Nora sits next to Owen, and Sam is not at either end.
    return abs(seat["Nora"] - seat["Owen"]) == 1 and seat["Sam"] not in (1, len(row))

def solve(options):
    return match(options, len([row for row in permutations(FRIENDS) if allowed(row)]))
