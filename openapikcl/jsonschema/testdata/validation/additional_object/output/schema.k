schema Schema:
    version?: str
    name?: str
    [...str]: SchemaAdditionalProps


schema SchemaAdditionalProps:
    enabled: bool
    priority?: int

    check:
        priority <= 5 if priority is not Undefined, "priority must be at most 5"
        priority >= 1 if priority is not Undefined, "priority must be at least 1"


