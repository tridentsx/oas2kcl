import regex

schema Person:
    age: int
    email: str?
    address: {str:any}?
    phoneNumbers: [str]?
    firstName: str
    lastName: str
