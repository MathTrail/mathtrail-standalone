def digits_up_to(last_page):
    return sum([len(str(page)) for page in range(1, last_page + 1)])

def solve(options):
    for pages in range(1, 200):
        if digits_up_to(pages) == 39:
            return match(options, pages)
    fail("no book of under 200 pages takes 39 digits to number")
