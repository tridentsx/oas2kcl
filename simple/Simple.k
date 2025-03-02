schema Simple:
    # Unique identifier
    id: str
    # Name of the entity
    name: str
    status: str | None
    tags: [str] | None

    check:
            len(name) >= 1
            len(name) <= 100
            status in ["active", "inactive", "pending"]
            len(tags) == len(set(tags))
