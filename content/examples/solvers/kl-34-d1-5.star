def says_yes(knight):
    # "Are you a knight?" is true exactly when a knight is asked, and a liar
    # turns the true answer round.
    true_answer = knight
    return true_answer if knight else not true_answer

def solve(options):
    answers = {
        (True, False): "Knights say yes, liars say no",
        (False, True): "Knights say no, liars say yes",
        (True, True): "Everybody says yes",
        (False, False): "Everybody says no",
    }
    return match(options, answers[(says_yes(True), says_yes(False))])
