NAMES = ["Olly", "Pam", "Rex", "Sue"]

def named(people):
    # A group as the options write it.
    if len(people) == 0:
        return "Nobody"
    if len(people) == 1:
        return "Only " + people[0]
    return ", ".join(people[:-1]) + " and " + people[-1]

def solve(options):
    tables = []
    for olly, pam, rex, sue in product([True, False], repeat=4):  # True is a knight
        olly_says = rex == olly  # 'Rex and I are both knights or both liars.'
        pam_says = olly != sue  # 'One of Olly and Sue is a knight and the other is a liar.'
        rex_says = pam or sue  # 'At least one of Pam and Sue is a knight.'
        sue_says = len([kind for kind in [olly, pam, rex, sue] if kind]) == 3  # 'Exactly three of us four are knights.'
        if olly_says == olly and pam_says == pam and rex_says == rex and sue_says == sue:
            tables.append({"Olly": olly, "Pam": pam, "Rex": rex, "Sue": sue})
    if len(tables) == 0:
        fail("no mix of knights and liars can say all of this")
    # Certainly a knight is a knight in every mix that fits.
    certain = [name for name in NAMES if all([table[name] for table in tables])]
    return match(options, named(certain))
