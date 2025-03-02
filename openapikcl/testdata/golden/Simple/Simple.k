# Unique identifier
schema Simple:
    # Unique identifier
    id: str
    # Name of the entity
    name: str
    tags: [str] | None
    status: str | None

    check:
        len(name) >= 1
        len(name) <= 100
        len(tags) == len(set(tags))
        status in ["active", "inactive", "pending"]