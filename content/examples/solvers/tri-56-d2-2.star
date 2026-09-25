RESULT = 10  # what Tom got at the end

def fits(number):
    # Times 3, add 7, halve, take 4 away; keep to whole numbers where the story halves.
    after = number * 3 + 7
    return after % 2 == 0 and after // 2 - 4 == RESULT

def solve(options):
    fitting = [n for n in range(1001) if fits(n)]
    if len(fitting) != 1:
        fail("%d whole numbers fit the story" % len(fitting))
    return match(options, fitting[0])
