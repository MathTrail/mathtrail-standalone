BREADS = ["white", "brown"]
FILLINGS = ["cheese", "ham", "egg"]
SAUCES = ["first sauce", "second sauce", "no sauce"]  # no sauce at all is a choice too

def solve(options):
    # Independent choices are a product.
    return match(options, len(product(BREADS, FILLINGS, SAUCES)))
