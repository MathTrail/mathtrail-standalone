def solve(options):
    # Every school of up to 400 whole pupils in which a quarter of those
    # outside the club and 40% of all are whole pupils.
    answers = set()
    for pupils in range(5, 401, 5):
        for outside in range(0, pupils + 1, 4):
            club = pupils - outside
            if (club + outside // 4) * 100 == 40 * pupils and club * 100 % pupils == 0:
                answers.add(club * 100 // pupils)
    if len(answers) != 1:
        fail("the schools give %d answers" % len(answers))
    return match(options, "%d%%" % list(answers)[0])
