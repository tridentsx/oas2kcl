# An order in the pet store
schema Order:
    id: int | None
    petId: int | None
    quantity: int | None
    shipDate: datetime | None
    # Order Status
    status: str | None
    complete: bool | None

    check:
        status in ["placed", "approved", "delivered"]
