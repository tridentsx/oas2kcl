
# Basic validation test 
# Create an instance of Helmvalues schema
instance = {
    name = "Test User"
    age = 30
    isActive = True
    status = "active"
}

# Validate it against the Helmvalues schema
check_instance = Helmvalues {
    name = instance.name
    age = instance.age
    isActive = instance.isActive
    status = instance.status
}
