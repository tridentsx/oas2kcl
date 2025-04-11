schema Book:
    price?: float
    available?: bool
    category?: str
    rule?: str
    name?: str
    castingOption?: str
    title: str
    authors?: [str]

    check:
        category in ["Fiction", "Science", "History"] if category is not Undefined 
        castingOption in ["originalName", "snakeCase", "camelCase"] if castingOption is not Undefined 


