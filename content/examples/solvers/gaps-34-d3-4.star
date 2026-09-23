def solve(options):
    cut_marks = range(10, 100, 10)
    return match(options, len(cut_marks) * 3)
