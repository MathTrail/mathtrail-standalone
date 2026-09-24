def solve(options):
    # Pages 40 to 100, both counted, kept when the number divides by 4.
    pictures = [page for page in range(40, 101) if page % 4 == 0]
    return match(options, len(pictures))
