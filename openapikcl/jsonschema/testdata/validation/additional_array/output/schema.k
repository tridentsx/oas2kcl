schema Config:
    name?: str
    [...str]: [ConfigAdditionalProps_items]


schema ConfigAdditionalProps_items:
    id: str
    value: int


