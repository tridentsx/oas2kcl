schema Category:
    # Category name
    name: str
    parentCategory: {"id": int | None, "name": str | None, "parentCategory": Category | None} | None
    # Category ID
    id: int
