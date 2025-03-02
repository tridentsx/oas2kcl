schema Order:
    # Order Status
    status: str | None
    # Default: false
    complete: bool | None
    id: int | None
    petId: int | None
    quantity: int | None
    shipDate: datetime | None

    check:
            status in ["placed", "approved", "delivered"]
