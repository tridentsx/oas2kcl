# No imports needed

schema Order:

    petId?: int
    status?: str
    id?: int

    check:
        status in ["placed", "approved", "delivered"]