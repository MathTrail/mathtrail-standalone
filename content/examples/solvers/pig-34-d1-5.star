def solve(options):
    children = 13
    months = 12
    # The fullest month is smallest when the birthdays are spread as evenly as
    # they go, so that is the case to decide the question on.
    for most in range(children + 1):
        if months * most >= children:
            return match(options, "Yes, always" if most >= 2 else "No, never")
    fail("no number of children in a month is ever enough")
