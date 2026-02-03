#!/usr/bin/env python3
"""
Script to configure Keycloak service account permissions for user management.
This grants the expense-app service account the ability to create and manage users.
"""

import urllib.request
import json
import time

KEYCLOAK_URL = "http://localhost:8080"
ADMIN_USER = "admin"
ADMIN_PASSWORD = "admin"
REALM = "expense-sharing"
CLIENT_ID = "expense-app"

def get_admin_token():
    """Get admin access token"""
    url = f"{KEYCLOAK_URL}/realms/master/protocol/openid-connect/token"
    data = {
        "client_id": "admin-cli",
        "username": ADMIN_USER,
        "password": ADMIN_PASSWORD,
        "grant_type": "password"
    }
    
    req = urllib.request.Request(
        url,
        data=urllib.parse.urlencode(data).encode(),
        headers={"Content-Type": "application/x-www-form-urlencoded"}
    )
    
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())["access_token"]

def get_service_account_user(token, client_uuid):
    """Get the service account user ID for the client"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients/{client_uuid}/service-account-user"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())["id"]

def get_client_uuid(token):
    """Get the UUID of the expense-app client"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients?clientId={CLIENT_ID}"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    
    with urllib.request.urlopen(req) as response:
        clients = json.loads(response.read())
        if clients:
            return clients[0]["id"]
    return None

def get_realm_management_client_uuid(token):
    """Get the UUID of the realm-management client"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients?clientId=realm-management"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    
    with urllib.request.urlopen(req) as response:
        clients = json.loads(response.read())
        if clients:
            return clients[0]["id"]
    return None

def get_role(token, realm_mgmt_uuid, role_name):
    """Get a role from realm-management client"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/clients/{realm_mgmt_uuid}/roles/{role_name}"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    
    try:
        with urllib.request.urlopen(req) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError:
        return None

def assign_roles_to_service_account(token, service_account_id, realm_mgmt_uuid, roles):
    """Assign multiple roles to the service account"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/users/{service_account_id}/role-mappings/clients/{realm_mgmt_uuid}"
    
    req = urllib.request.Request(
        url,
        data=json.dumps(roles).encode(),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        },
        method="POST"
    )
    
    try:
        with urllib.request.urlopen(req) as response:
            return response.status == 204
    except urllib.error.HTTPError as e:
        print(f"Warning: {e.code} - {e.read().decode()}")
        return False

def main():
    print("Configuring Keycloak service account permissions...")
    
    # Wait for Keycloak to be ready
    print("Waiting for Keycloak to be ready...")
    time.sleep(5)
    
    try:
        # Get admin token
        print("Getting admin token...")
        token = get_admin_token()
        
        # Get expense-app client UUID
        print(f"Finding {CLIENT_ID} client...")
        client_uuid = get_client_uuid(token)
        if not client_uuid:
            print(f"ERROR: Client {CLIENT_ID} not found!")
            return False
        print(f"Found client: {client_uuid}")
        
        # Get service account user
        print("Getting service account user...")
        service_account_id = get_service_account_user(token, client_uuid)
        print(f"Service account user ID: {service_account_id}")
        
        # Get realm-management client
        print("Finding realm-management client...")
        realm_mgmt_uuid = get_realm_management_client_uuid(token)
        if not realm_mgmt_uuid:
            print("ERROR: realm-management client not found!")
            return False
        print(f"Found realm-management: {realm_mgmt_uuid}")
        
        # Get all required roles
        print("Getting required roles...")
        required_role_names = [
            "manage-users",
            "view-users",
            "query-users",
            "create-client"
        ]
        
        roles_to_assign = []
        for role_name in required_role_names:
            role = get_role(token, realm_mgmt_uuid, role_name)
            if role:
                print(f"  ✓ Found role: {role_name}")
                roles_to_assign.append(role)
            else:
                print(f"  ✗ Role not found: {role_name}")
        
        if not roles_to_assign:
            print("ERROR: No roles found to assign!")
            return False
        
        # Assign all roles to service account
        print(f"Assigning {len(roles_to_assign)} roles to service account...")
        assign_roles_to_service_account(token, service_account_id, realm_mgmt_uuid, roles_to_assign)
        
        print("✅ Successfully configured service account permissions!")
        print(f"The {CLIENT_ID} service account now has {len(roles_to_assign)} roles.")
        return True
        
    except urllib.error.HTTPError as e:
        print(f"HTTP Error: {e.code} - {e.read().decode()}")
        return False
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    success = main()
    exit(0 if success else 1)
