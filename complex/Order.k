schema Order:
    # Order ID
    id: str
    items: [{"product": Product | None, "quantity": int | None, "unitPrice": float | None}]
    # Default: pending
    status: str | None
    # Total order amount
    totalAmount: float | None
    customer: any

    check:
            len(items) >= 1
            status in ["pending", "processing", "shipped", "delivered", "cancelled"]
