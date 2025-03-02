# Base object with common fields
schema BaseObject:
    # Unique identifier
    id: str
    # Creation timestamp
    createdAt: datetime | None
    # Last update timestamp
    updatedAt: datetime | None
