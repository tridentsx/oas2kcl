import regex

schema Order:

    status?: str
    # Total order amount
    totalAmount?: float
    customer: Customer
    # Order ID
    id: str
    items: [OrderItem]

    check:
        status in ["pending", "processing", "shipped", "delivered", "cancelled"]
        regex.match(id, r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
        len(items) >= 1