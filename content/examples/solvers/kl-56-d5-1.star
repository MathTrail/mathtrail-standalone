def solve(options):
    # Every number of knights, and every count of knights and of liars on
    # blue chairs before the change and after it. At first every 'blue' was
    # true for a knight and false for a liar; after the change, 'red' is said
    # by the knights on red chairs and by the liars on blue ones.
    answers = set()
    for knights in range(0, 25):
        liars = 24 - knights
        for knights_first in range(0, knights + 1):
            for liars_first in range(0, liars + 1):
                if knights_first != knights or liars_first != 0:
                    continue
                blue = knights_first + liars_first
                for knights_on_blue in range(0, knights + 1):
                    for liars_on_blue in range(0, liars + 1):
                        if knights_on_blue + liars_on_blue != blue:
                            continue
                        if (knights - knights_on_blue) + liars_on_blue == 8:
                            answers.add(knights - knights_on_blue)
    if not answers:
        fail("no seating fits")
    if len(answers) != 1:
        return match(options, "It cannot be found without the number of knights")
    return match(options, list(answers)[0])
