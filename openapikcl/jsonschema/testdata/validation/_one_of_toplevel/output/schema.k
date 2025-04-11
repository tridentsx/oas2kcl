


schema Pet_oneOf_0:
    bark?: bool
    breed: str
    name: str

    check:
        breed in ["dog"] if breed is not Undefined 


schema Pet_oneOf_1:
    meow?: bool
    breed: str
    name: str

    check:
        breed in ["cat"] if breed is not Undefined 


