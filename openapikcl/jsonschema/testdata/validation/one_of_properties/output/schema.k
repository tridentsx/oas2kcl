schema Book:
    category?: Book_category_oneOf_0 | Book_category_oneOf_1 
    title?: str
    author?: str | [str] | int 


schema Book_category_oneOf_0:
    name?: str


schema Book_category_oneOf_1:
    title?: str


