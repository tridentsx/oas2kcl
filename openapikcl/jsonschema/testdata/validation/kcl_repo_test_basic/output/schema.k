schema Book:
    rule?: str
    title: str
    authors?: [str]
    price?: float
    available?: bool
    category?: str

    check:
        category in ["Fiction", "Science", "History"] if category is not Undefined 


