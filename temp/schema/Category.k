import regex

schema Category:
    # Category ID
    id: int
    # Category name
    name: str
    # Circular reference to Category
    parentCategory?: dict