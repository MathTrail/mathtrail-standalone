# For quantities linked by "a for every b": start from the one the question
# gives, find each next quantity from the one before it, one link at a time,
# and read the answer off the whole chain.
# From reference task ratio-56-d4-1.

START = 30  # the quantity the question gives: here, the daisies
LINKS = [(4, 5), (2, 3)]  # each link: a of the next quantity for every b of the one before
LARGEST = 1000  # the largest count worth trying

def next_count(count, a, b):
    # The one count that is a for every b of this count.
    found = [n for n in range(0, LARGEST + 1) if n * b == count * a]
    if len(found) != 1:
        fail("%d counts are %d for every %d of %d" % (len(found), a, b, count))
    return found[0]

def solve(options):
    chain = [START]
    for a, b in LINKS:
        chain.append(next_count(chain[-1], a, b))
    return match(options, sum(chain))  # chain[-1] when the question asks for the last quantity alone
