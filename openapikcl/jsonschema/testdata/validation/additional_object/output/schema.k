schema Schema:
    name?: str
    version?: str
    [...str]: SchemaAdditionalProps


schema SchemaAdditionalProps:
    priority?: int
    enabled: bool

    check:
        priority >= 1 if priority is not Undefined, "priority must be at least 1"
        priority <= 5 if priority is not Undefined, "priority must be at most 5"


