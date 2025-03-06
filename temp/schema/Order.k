import regex
import Address
import BaseObject
import Category
import Product

schema Order:
    customer: any
    # Order ID
    id: str
    items: [dict]
    status?: str
    # Total order amount
    totalAmount?: float

    check:
        len(items) >= 1
        status in ["pending", "processing", "shipped", "delivered", "cancelled"]