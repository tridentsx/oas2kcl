schema Book:
    authors?: [str]
    price?: float
    available?: bool
    category?: str
    rule?: str
    title: str

    check:
        category in ["Fiction", "Science", "History"] if category is not Undefined 


