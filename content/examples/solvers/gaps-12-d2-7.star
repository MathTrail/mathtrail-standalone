def solve(options, picture):
    posts = 0
    for item in picture["items"]:
        if not item.get("skip", False):
            posts += 1
    rails = int(picture["gaps"])
    return match(options, (posts - 1) * rails)
