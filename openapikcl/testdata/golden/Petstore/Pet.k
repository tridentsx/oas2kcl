# A pet in the store
schema Pet:
    id: int | None
    name: str
    category: Category | None
    photoUrls: [str]
    tags: [Tag] | None
    # pet status in the store
    status: str | None

    check:
        status in ["available", "pending", "sold"]
