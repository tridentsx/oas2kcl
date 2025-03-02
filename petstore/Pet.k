schema Pet:
    # Default: 40
    # ReadOnly: This field is read-only
    id: int | None
    name: str
    photoUrls: [str]
    # pet status in the store
    status: str | None
    tags: [{"name": str | None, "id": int | None}] | None
    category: {"id": int | None, "name": str | None} | None

    check:
            status in ["available", "pending", "sold"]
