def solve(options):
    # Whether each sentence is true when a knight says it and when a liar does.
    truth = {
        "I am a knight.": (True, False),
        "I am a liar.": (False, True),
        "2 + 2 = 4.": (True, True),
        "2 + 2 = 5.": (False, False),
        "I am not a liar.": (True, False),
    }
    # A knight can say only what is then true, and a liar only what is then false.
    unsayable = [text for text in truth if not truth[text][0] and truth[text][1]]
    if len(unsayable) != 1:
        fail("%d of the sentences can never be said" % len(unsayable))
    return match(options, unsayable[0])
