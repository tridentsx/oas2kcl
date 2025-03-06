import regex

schema Address:
    # City
    city: str
    # Country
    country: str
    # Postal code
    postalCode?: str
    # State or province
    state?: str
    # Street address
    street: str