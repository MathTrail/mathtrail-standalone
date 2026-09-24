def named(people):
    # A group as the options write it.
    if len(people) == 0:
        return "Nobody"
    if len(people) == 1:
        return "Only " + people[0]
    return ", ".join(people[:-1]) + " and " + people[-1]

def solve(options):
    answers = set()
    for tom, uma, val, wes in product([True, False], repeat=4):  # True is a knight
        tom_says = not uma  # 'Uma is lying.'
        uma_says = not val  # 'Val is lying.'
        val_says = not tom and not uma  # 'Tom and Uma are both lying.'
        wes_says = len([kind for kind in [tom, uma, val] if kind]) == 1  # 'Exactly one of Tom, Uma and Val is telling the truth.'
        if tom_says == tom and uma_says == uma and val_says == val and wes_says == wes:
            kinds = {"Tom": tom, "Uma": uma, "Val": val, "Wes": wes}
            answers.add(named([name for name in ["Tom", "Uma", "Val", "Wes"] if kinds[name]]))  # the knights
    if len(answers) != 1:
        fail("the words fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
