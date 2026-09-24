CANNOT_TELL = "It is impossible to tell"

def named(people):
    # A group as the options write it.
    if len(people) == 0:
        return "Nobody"
    if len(people) == 1:
        return "Only " + people[0]
    return ", ".join(people[:-1]) + " and " + people[-1]

def verdict(answers):
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    answers = set()
    for lou, meg, ned in product([True, False], repeat=3):  # True is a knight
        # Meg would say 'Ned is a liar' exactly when saying it is what her kind does:
        # a knight says it when it is true, a liar when it is false.
        lou_says = (not ned) == meg  # 'If you asked Meg, she would say that Ned is a liar.'
        meg_says = ned  # 'Ned is a knight.'
        ned_says = not lou  # 'Lou is a liar.'
        if lou_says == lou and meg_says == meg and ned_says == ned:
            kinds = {"Lou": lou, "Meg": meg, "Ned": ned}
            answers.add(named([name for name in ["Lou", "Meg", "Ned"] if not kinds[name]]))  # the liars
    return match(options, verdict(answers))
