def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    names = {
        (): "nobody",
        ("Ann",): "only Ann",
        ("Ben",): "only Ben",
        ("Kim",): "only Kim",
        ("Ann", "Kim"): "Ann and Kim",
    }
    answers = set()
    for ann, ben, kim in product([True, False], repeat=3):
        ann_says = not ben  # 'Ben is a liar.'
        ben_says = not ann and not kim  # 'Ann and Kim are both liars.'
        kim_says = ben  # 'Ben is a knight.'
        if ann_says == ann and ben_says == ben and kim_says == kim:
            knights = tuple([name for name, kind in [("Ann", ann), ("Ben", ben), ("Kim", kim)] if kind])
            answers.add(names.get(knights, "some other mix"))
    return match(options, verdict(answers))
