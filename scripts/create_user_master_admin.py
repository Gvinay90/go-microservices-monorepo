#!/usr/bin/env python3
"""Test creating user with master admin token (full permissions)"""

import urllib.request
import json

def main():
    print("Testing user creation with MASTER ADMIN token...")
    
    # 1. Get master admin token
    token_url = "http://localhost:8080/realms/master/protocol/openid-connect/token"
    token_data = urllib.parse.urlencode({
        "client_id": "admin-cli",
        "username": "admin",
        "password": "admin",
        "grant_type": "password"
    }).encode()
    
    token_req = urllib.request.Request(token_url, data=token_data, headers={"Content-Type": "application/x-www-form-urlencoded"})
    
    with urllib.request.urlopen(token_req) as response:
        token = json.loads(response.read())["access_token"]
        print("✅ Got admin token")
    
    # 2. Create user
    user_url = "http://localhost:8080/admin/realms/expense-sharing/users"
    user_data = {
        "username": "vinay@email.com",
        "email": "vinay@email.com",
        "enabled": True,
        "emailVerified": True,
        "firstName": "Vinay",
        "lastName": "Gupta",
        "credentials": [{
            "type": "password",
            "value": "123456789",
            "temporary": False
        }]
    }
    
    user_req = urllib.request.Request(
        user_url,
        data=json.dumps(user_data).encode(),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        },
        method="POST"
    )
    
    try:
        with urllib.request.urlopen(user_req) as response:
            print(f"✅ User created! Status: {response.status}")
            location = response.headers.get("Location")
            if location:
                print(f"User ID: {location.split('/')[-1]}")
            return True
    except urllib.error.HTTPError as e:
        print(f"❌ Failed: {e.code} - {e.read().decode()}")
        return False

if __name__ == "__main__":
    import sys
    sys.exit(0 if main() else 1)
