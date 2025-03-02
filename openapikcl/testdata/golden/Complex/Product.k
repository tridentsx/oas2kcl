schema Product:
    mixin [BaseObject]
    
    # Product name
    name: str
    # Product price
    price: float
    category: Category | None
    # Product tags
    tags: [str] | None
    # Additional metadata
    metadata: {str: str} | None
    status: str | None

    check:
        len(name) >= 1
        len(name) <= 100
        price >= 0
        status in ["available", "out_of_stock", "discontinued"]
