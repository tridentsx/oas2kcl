schema Product:
    # Product name
    name: str
    category: {"parentCategory": Category | None, "id": int | None, "name": str | None} | None
    # Product price
    price: float
    # Creation timestamp
    createdAt: datetime | None
    # Last update timestamp
    updatedAt: datetime | None
    # Additional metadata
    metadata: dict | None
    # Default: available
    status: str | None
    # Product tags
    tags: [str] | None
    # Unique identifier
    id: str

    check:
            len(name) >= 1
            len(name) <= 100
            price >= 0
            status in ["available", "out_of_stock", "discontinued"]
