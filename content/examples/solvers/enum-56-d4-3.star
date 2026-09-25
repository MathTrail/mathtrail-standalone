PUPILS = ["Ella", "Finn", "Gia", "Hugo", "Ivy", "Jack"]

def allowed(team):
    # Gia is in the team, and Ella and Finn are not both in it.
    return "Gia" in team and not ("Ella" in team and "Finn" in team)

def solve(options):
    # A team is a group, so the order of picking does not matter.
    return match(options, len([team for team in combinations(PUPILS, 3) if allowed(team)]))
