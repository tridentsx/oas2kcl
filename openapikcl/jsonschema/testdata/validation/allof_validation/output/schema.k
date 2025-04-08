schema Schema:
    price?: float

    check:
        price >= 0 if price is not Undefined, "price must be at least 0"


