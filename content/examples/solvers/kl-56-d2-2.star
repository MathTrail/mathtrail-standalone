def named(people):
    # A group as the options write it.
    if len(people) == 0:
        return "Nobody"
    if len(people) == 1:
        return "Only " + people[0]
    return ", ".join(people[:-1]) + " and " + people[-1]

def solve(options):
    answers = set()
    for fay, gus, hal in product([True, False], repeat=3):  # True is a knight
        fay_says = not (fay and gus and hal)  # 'At least one of us three is a liar.'
        gus_says = not fay  # 'Fay is a liar.'
        hal_says = not gus  # 'Gus is a liar.'
        if fay_says == fay and gus_says == gus and hal_says == hal:
            kinds = {"Fay": fay, "Gus": gus, "Hal": hal}
            answers.add(named([name for name in ["Fay", "Gus", "Hal"] if not kinds[name]]))  # the liars
    if len(answers) != 1:
        fail("the words fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
