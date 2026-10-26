SENTENCES = {
    "'Exactly three of you are knights.'": lambda you, us: you == 3,
    "'Exactly four of us are knights.'": lambda you, us: us == 4,
    "'Exactly four of you are knights.'": lambda you, us: you == 4,
    "'All of you are liars.'": lambda you, us: you == 0,
    "'At least one of you is a liar.'": lambda you, us: you < 5,
}

def solve(options):
    # Four knights and two liars: a knight says only what is true for it,
    # and a liar only what is false; "you" are the other five.
    people = [True, True, True, True, False, False]
    sayable = []
    for sentence, holds in SENTENCES.items():
        us = len([k for k in people if k])
        if all([holds(len([k for j, k in enumerate(people) if j != i and k]), us) == knight for i, knight in enumerate(people)]):
            sayable.append(sentence)
    if len(sayable) != 1:
        fail("%d sentences could be said by all" % len(sayable))
    return match(options, sayable[0])
