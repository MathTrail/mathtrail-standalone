def solve(options, picture):
    on = set(picture["cells"])
    for mark in picture["marks"]:
        if mark["at"] in on:
            return match(options, 7 - int(mark["label"]))
    return []
