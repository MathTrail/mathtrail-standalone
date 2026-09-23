def solve(options):
    posts = [2 * i for i in range(8)]
    return match(options, posts[-1] - posts[0])
