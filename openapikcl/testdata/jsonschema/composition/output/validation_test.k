
# Basic validation test 
# Create an instance of CompositionTest schema
instance = {
    name = "Test User"
    age = 30
    isActive = True
    status = "active"
}

# Validate it against the CompositionTest schema
check_instance = CompositionTest {
    name = instance.name
    age = instance.age
    isActive = instance.isActive
    status = instance.status
}
