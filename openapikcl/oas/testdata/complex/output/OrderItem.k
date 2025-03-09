import regex

schema OrderItem:

    product: Product
    # Item quantity
    quantity: int
    # Price per unit
    unitPrice: float

    check:
        quantity >= 1