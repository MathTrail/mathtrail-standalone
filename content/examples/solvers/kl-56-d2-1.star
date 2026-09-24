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
    for ada, ben, cal in product([True, False], repeat=3):  # True is a knight
        ada_says = not ben  # 'Ben is a liar.'
        ben_says = cal  # 'Cal is a knight.'
        cal_says = ada == cal  # 'Ada and I are both knights or both liars.'
        if ada_says == ada and ben_says == ben and cal_says == cal:
            kinds = {"Ada": ada, "Ben": ben, "Cal": cal}
            answers.add(named([name for name in ["Ada", "Ben", "Cal"] if kinds[name]]))  # the knights
    return match(options, verdict(answers))
