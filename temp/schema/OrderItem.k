import regex
import BaseObject
import Category

schema OrderItem:
    product: any
    # Item quantity
    quantity: int
    # Price per unit
    unitPrice: float

    check:
        quantity >= 1