schema Pet_allOf_0Mixin:
    name: str
    age?: int


schema Pet_allOf_1Mixin:
    breed: str


schema Pet:
    mixin [Pet_allOf_0Mixin, Pet_allOf_1Mixin]


