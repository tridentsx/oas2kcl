# No imports needed

schema Order:

    id?: int
    petId?: int
    status?: str

    check:
        status in ["placed", "approved", "delivered"]