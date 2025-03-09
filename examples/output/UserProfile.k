import regex

schema UserProfile:
    # A user profile with various string constraints
    password: str  # Password (must contain at least one uppercase letter, one lowercase letter, one number, and one special character)
    bio?: str  # User's biography
    addresses?: list[UserProfileAddress]  # User's addresses
    displayName?: str  # User's display name
    roles?: list[str]  # User roles
    username: str  # Username (alphanumeric and underscores only)
    email: str  # Email address
    website?: str  # User's website URL
    birthdate?: str  # User's birthdate (YYYY-MM-DD)
    lastLogin?: str  # User's last login timestamp
    status?: str  # Account status
    tags?: list[str]  # User tags (alphanumeric and hyphens only)
    emails?: list[str]  # Additional email addresses
    phone?: str  # Phone number in E.164 format
    userId: str  # User's unique identifier (UUID)
    socialProfiles?: UserProfileSocialProfiles  # Social profiles

schema UserProfileAddress:
    line1: str  # First line of address
    line2?: str  # Second line of address
    city: str  # City
    postalCode: str  # Postal code
    country: str  # Country code (2-letter ISO code)

schema UserProfileSocialProfiles:
    twitter?: str  # Twitter handle
    facebook?: str  # Facebook profile URL
    linkedin?: str  # LinkedIn profile URL
    github?: str  # GitHub username
