import regex

schema UserProfile:
    # A user profile with various string constraints
    username: str  # Username (alphanumeric and underscores only)
    password: str  # Password
    email: str     # Email address
    userId: str    # User's unique identifier

schema UserProfileAddress:
    line1: str     # First line of address
    city: str      # City
    postalCode: str # Postal code
    country: str   # Country code (2-letter ISO code)

schema UserProfileSocialProfiles:
    twitter: str   # Twitter handle
    facebook: str  # Facebook profile URL
    linkedin: str  # LinkedIn profile URL
    github: str    # GitHub username

# Validator schema to demonstrate validations
schema UserProfileValidator:
    profile: UserProfile
    
    check:
        len(profile.username) >= 3, "Username must be at least 3 characters"
        len(profile.username) <= 20, "Username must be at most 20 characters"
        regex.match("^[a-zA-Z0-9_]+$", profile.username), "Username must only contain alphanumeric characters and underscores"
        
        len(profile.password) >= 8, "Password must be at least 8 characters"
        regex.match("^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$", profile.password), "Password must meet complexity requirements"
        
        regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$", profile.email), "Email must be valid format"