import regex
import Category

schema Product:
    # Circular reference to Product
    category?: dict
    # Creation timestamp
    createdAt?: str
    # Unique identifier
    id: str
    # Additional metadata
    metadata?: dict
    # Product name
    name: str
    # Product price
    price: float
    status?: str
    # Product tags
    tags?: [str]
    # Last update timestamp
    updatedAt?: str

    check:
        len(name) >= 1
        len(name) <= 100
        price >= 0
        status in ["available", "out_of_stock", "discontinued"]