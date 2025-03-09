import regex

schema UserProfile:
    # A user profile with various string constraints
    # User roles
    roles?: list[str]
    # Unique items: true
    # Item allowed values: "admin", "moderator", "user", "guest"

    # User tags (alphanumeric and hyphens only)
    tags?: list[str]
    # Item min length: 1
    # Item max length: 20
    # Item pattern: ^[a-zA-Z0-9-]+$

    # User's birthdate (YYYY-MM-DD)
    birthdate?: str
    # Format: date

    # User's last login timestamp
    lastLogin?: str
    # Format: date-time

    # User's website URL
    website?: str
    # Format: uri

    # Account status
    status?: str
    # Allowed values: "active", "inactive", "suspended", "pending"

    # Username (alphanumeric and underscores only)
    username: str
    # Min length: 3
    # Max length: 20
    # Regex pattern: ^[a-zA-Z0-9_]+$

    # Phone number in E.164 format
    phone?: str
    # Regex pattern: ^\+?[1-9]\d{1,14}$

    # User's unique identifier (UUID)
    userId: str
    # Format: uuid

    # Additional email addresses
    emails?: list[str]
    # Unique items: true
    # Item format: email

    socialProfiles?: UserProfileSocialProfiles
    addresses?: list[UserProfileAddress]
    # Password (must contain at least one uppercase letter, one lowercase letter, one number, and one special character)
    password: str
    # Min length: 8
    # Max length: 100
    # Regex pattern: ^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$

    # Email address
    email: str
    # Format: email

    # User's display name
    displayName?: str
    # Min length: 1
    # Max length: 50

    # User's biography
    bio?: str
    # Max length: 500


schema UserProfileSocialProfiles:
    # Twitter handle
    twitter?: str
    # Facebook profile URL
    facebook?: str
    # LinkedIn profile URL
    linkedin?: str
    # GitHub username
    github?: str

schema UserProfileAddress:
    line1: str
    # Min length: 1
    # Max length: 100
    line2?: str
    # Max length: 100
    city: str
    # Min length: 1
    # Max length: 100
    postalCode: str
    # Regex pattern: ^[A-Z0-9]{1,10}$
    country: str
    # Min length: 2
    # Max length: 2
    # Regex pattern: ^[A-Z]{2}$

    if line1 == None:
        assert False, "line1 is required"
    if city == None:
        assert False, "city is required"
    if postalCode == None:
        assert False, "postalCode is required"
    if country == None:
        assert False, "country is required"