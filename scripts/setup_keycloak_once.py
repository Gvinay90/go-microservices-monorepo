#!/usr/bin/env python3
"""
One-time Keycloak setup: Assign service account roles to expense-app client.
This script automates the manual UI steps:
1. Go to Clients → expense-app
2. Click Service account roles tab
3. Assign realm-management roles (manage-users, view-users, create-client)
"""

import urllib.request
import json
import time
import sys

KEYCLOAK_URL = "http://localhost:8080"
REALM = "expense-sharing"

def get_admin_token():
    url = f"{KEYCLOAK_URL}/realms/master/protocol/openid-connect/token"
    data = urllib.parse.urlencode({
        "client_id": "admin-cli",
        "username": "admin",
        "password": "admin",
        "grant_type": "password"
    }).encode()
    
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/x-www-form-urlencoded"})
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())["access_token"]

def get_client_uuid(token, client_id):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients?clientId={client_id}"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as response:
        clients = json.loads(response.read())
        return clients[0]["id"] if clients else None

def get_service_account_user(token, client_uuid):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients/{client_uuid}/service-account-user"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())["id"]

def get_client_role(token, client_uuid, role_name):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients/{client_uuid}/roles/{role_name}"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(req) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError:
        return None

def assign_client_roles(token, user_id, client_uuid, roles):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/users/{user_id}/role-mappings/clients/{client_uuid}"
    req = urllib.request.Request(
        url,
        data=json.dumps(roles).encode(),
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
        method="POST"
    )
    with urllib.request.urlopen(req) as response:
        return response.status == 204

def main():
    print("=" * 70)
    print("Keycloak Service Account Setup")
    print("=" * 70)
    
    # Wait for Keycloak
    print("\nWaiting for Keycloak to be fully ready...")
    for i in range(30):
        try:
            urllib.request.urlopen(f"{KEYCLOAK_URL}/health/ready", timeout=2)
            print("✅ Keycloak is ready!")
            break
        except:
            time.sleep(2)
            print(f"  Waiting... ({i+1}/30)")
    else:
        print("❌ Keycloak failed to start")
        return False
    
    time.sleep(3)  # Extra wait for import to complete
    
    try:
        # Step 1: Get admin token
        print("\n1. Getting admin token...")
        token = get_admin_token()
        print("   ✓ Admin token acquired")
        
        # Step 2: Get expense-app client UUID
        print("\n2. Finding expense-app client...")
        expense_app_uuid = get_client_uuid(token, "expense-app")
        if not expense_app_uuid:
            print("   ❌ Client not found!")
            return False
        print(f"   ✓ Found: {expense_app_uuid}")
        
        # Step 3: Get service account user
        print("\n3. Getting service account user...")
        sa_user_id = get_service_account_user(token, expense_app_uuid)
        print(f"   ✓ Service account user: {sa_user_id}")
        
        # Step 4: Get realm-management client UUID
        print("\n4. Finding realm-management client...")
        realm_mgmt_uuid = get_client_uuid(token, "realm-management")
        if not realm_mgmt_uuid:
            print("   ❌ realm-management not found!")
            return False
        print(f"   ✓ Found: {realm_mgmt_uuid}")
        
        # Step 5: Get required roles
        print("\n5. Getting realm-management roles...")
        role_names = ["manage-users", "view-users", "create-client", "query-users"]
        roles = []
        for role_name in role_names:
            role = get_client_role(token, realm_mgmt_uuid, role_name)
            if role:
                roles.append(role)
                print(f"   ✓ {role_name}")
            else:
                print(f"   ✗ {role_name} (not found)")
        
        if not roles:
            print("\n❌ No roles found to assign!")
            return False
        
        # Step 6: Assign roles
        print(f"\n6. Assigning {len(roles)} roles to service account...")
        assign_client_roles(token, sa_user_id, realm_mgmt_uuid, roles)
        print("   ✓ Roles assigned successfully!")
        
        print("\n" + "=" * 70)
        print("✅ SETUP COMPLETE!")
        print("=" * 70)
        print("\nThe expense-app service account now has permission to:")
        for role in roles:
            print(f"  • {role['name']}")
        print("\nYou can now register new users through the app!")
        
        return True
        
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
