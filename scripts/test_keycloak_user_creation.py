#!/usr/bin/env python3
"""Test Keycloak user creation with service account"""

import urllib.request
import json

KEYCLOAK_URL = "http://localhost:8080"
REALM = "expense-sharing"
CLIENT_ID = "expense-app"
CLIENT_SECRET = "1aLlWaA8A37Qa6HyyUGdbx9tnGZRmXAH"  # From realm-export.json

def get_token():
    """Get service account token"""
    url = f"{KEYCLOAK_URL}/realms/{REALM}/protocol/openid-connect/token"
    data = urllib.parse.urlencode({
        "client_id": CLIENT_ID,
        "client_secret": CLIENT_SECRET,
        "grant_type": "client_credentials"
    }).encode()
    
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/x-www-form-urlencoded"})
    
    try:
        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read())
            print(f"✅ Token acquired successfully!")
            print(f"Token type: {result.get('token_type')}")
            print(f"Expires in: {result.get('expires_in')} seconds")
            return result["access_token"]
    except urllib.error.HTTPError as e:
        print(f"❌ Failed to get token: {e.code}")
        print(e.read().decode())
        return None

def create_test_user(token):
    """Try to create a test user"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/users"
    user_data = {
        "username": "testuser@example.com",
        "email": "testuser@example.com",
        "enabled": True,
        "emailVerified": True,
        "firstName": "Test",
        "credentials": [{
            "type": "password",
            "value": "testpass123",
            "temporary": False
        }]
    }
    
    req = urllib.request.Request(
        url,
        data=json.dumps(user_data).encode(),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        },
        method="POST"
    )
    
    try:
        with urllib.request.urlopen(req) as response:
            print(f"✅ User created successfully! Status: {response.status}")
            location = response.headers.get("Location")
            if location:
                user_id = location.split("/")[-1]
                print(f"User ID: {user_id}")
            return True
    except urllib.error.HTTPError as e:
        print(f"❌ Failed to create user: {e.code}")
        error_body = e.read().decode()
        print(f"Error details: {error_body}")
        return False

def main():
    print("Testing Keycloak user creation with service account...")
    print("=" * 60)
    
    # Step 1: Get token
    print("\n1. Getting service account token...")
    token = get_token()
    if not token:
        print("\n❌ Cannot proceed without a valid token")
        return False
    
    # Step 2: Try to create user
    print("\n2. Attempting to create test user...")
    success = create_test_user(token)
    
    print("\n" + "=" * 60)
    if success:
        print("✅ ALL TESTS PASSED!")
    else:
        print("❌ TEST FAILED - Check permissions")
    
    return success

if __name__ == "__main__":
    import sys
    success = main()
    sys.exit(0 if success else 1)
