schema Config:
    name?: str
    [...str]: [ConfigAdditionalProps_items]


schema ConfigAdditionalProps_items:
    value: int
    id: str


