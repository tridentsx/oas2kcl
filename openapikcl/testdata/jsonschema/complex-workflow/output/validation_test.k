
# Basic validation test 
# Create an instance of Schema schema
instance = {
    name = "Test User"
    age = 30
    isActive = True
    status = "active"
}

# Validate it against the Schema schema
check_instance = Schema {
    name = instance.name
    age = instance.age
    isActive = instance.isActive
    status = instance.status
}
