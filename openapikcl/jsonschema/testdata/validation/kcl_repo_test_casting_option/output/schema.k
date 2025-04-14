schema Book:
    castingOption?: str
    title: str
    authors?: [str]
    price?: float
    available?: bool
    category?: str
    rule?: str
    name?: str

    check:
        castingOption in ["originalName", "snakeCase", "camelCase"] if castingOption is not Undefined 
        category in ["Fiction", "Science", "History"] if category is not Undefined 


