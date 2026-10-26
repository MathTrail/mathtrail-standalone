QUESTIONS = {
    "'How many times have I asked you this question?'": lambda asked: asked,
    "'Do you always tell the truth?'": lambda asked: "yes",
    "'How much is 7 + 8?'": lambda asked: 7 + 8,
    "'What is your name?'": lambda asked: "Ilya",
}

def solve(options):
    # Ilya's true answer to each question the first time it is asked and the
    # second: he tells the truth both times, so his answers differ only where
    # the true answer does.
    changing = [question for question, answer in QUESTIONS.items() if answer(1) != answer(2)]
    if not changing:
        return match(options, "None: someone who always tells the truth never changes an answer")
    if len(changing) != 1:
        fail("%d questions could be answered differently" % len(changing))
    return match(options, changing[0])
