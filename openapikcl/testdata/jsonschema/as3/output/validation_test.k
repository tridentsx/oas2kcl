
# Basic validation test 
# Create an instance of AS3_API_Request schema
instance = {
    name = "Test User"
    age = 30
    isActive = True
    status = "active"
}

# Validate it against the AS3_API_Request schema
check_instance = AS3_API_Request {
    name = instance.name
    age = instance.age
    isActive = instance.isActive
    status = instance.status
}
