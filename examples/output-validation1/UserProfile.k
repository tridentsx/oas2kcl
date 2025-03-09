import regex

schema UserProfile:
    # A user profile with various string constraints
    # User's display name
    displayName?: str
    if displayName:
        assert len(displayName) >= 1, "displayName length must be at least 1"
        assert len(displayName) <= 50, "displayName length must be at most 50"

    # User's biography
    bio?: str
    if bio:
        assert len(bio) <= 500, "bio length must be at most 500"

    # User roles
    roles?: list[str]
    # Unique items: true
    # Item allowed values: "admin", "moderator", "user", "guest"

    socialProfiles?: UserProfileSocialProfiles
    # Username (alphanumeric and underscores only)
    username: str
    if username:
        assert len(username) >= 3, "username length must be at least 3"
        assert len(username) <= 20, "username length must be at most 20"
    # Regex pattern: ^[a-zA-Z0-9_]+$

    # Email address
    email: str
    # Format: email

    # User's website URL
    website?: str
    # Format: uri

    # Phone number in E.164 format
    phone?: str
    # Regex pattern: ^\+?[1-9]\d{1,14}$

    # User tags (alphanumeric and hyphens only)
    tags?: list[str]
    # Item min length: 1
    # Item max length: 20
    # Item pattern: ^[a-zA-Z0-9-]+$

    addresses?: list[UserProfileAddress]
    # Password (must contain at least one uppercase letter, one lowercase letter, one number, and one special character)
    password: str
    if password:
        assert len(password) >= 8, "password length must be at least 8"
        assert len(password) <= 100, "password length must be at most 100"
    # Regex pattern: ^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$

    # User's birthdate (YYYY-MM-DD)
    birthdate?: str
    # Format: date

    # User's last login timestamp
    lastLogin?: str
    # Format: date-time

    # User's unique identifier (UUID)
    userId: str
    # Format: uuid

    # Account status
    status?: str
    # Allowed values: "active", "inactive", "suspended", "pending"

    # Additional email addresses
    emails?: list[str]
    # Unique items: true
    # Item format: email


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
    if line1:
        assert len(line1) >= 1, "line1 length must be at least 1"
        assert len(line1) <= 100, "line1 length must be at most 100"
    line2?: str
    if line2:
        assert len(line2) <= 100, "line2 length must be at most 100"
    city: str
    if city:
        assert len(city) >= 1, "city length must be at least 1"
        assert len(city) <= 100, "city length must be at most 100"
    postalCode: str
    # Regex pattern: ^[A-Z0-9]{1,10}$
    country: str
    if country:
        assert len(country) >= 2, "country length must be at least 2"
        assert len(country) <= 2, "country length must be at most 2"
    # Regex pattern: ^[A-Z]{2}$

    if line1 == None:
        assert False, "line1 is required"
    if city == None:
        assert False, "city is required"
    if postalCode == None:
        assert False, "postalCode is required"
    if country == None:
        assert False, "country is required"