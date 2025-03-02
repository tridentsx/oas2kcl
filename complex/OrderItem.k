schema OrderItem:
    # Item quantity
    quantity: int
    # Price per unit
    unitPrice: float
    product: any

    check:
            quantity >= 1
