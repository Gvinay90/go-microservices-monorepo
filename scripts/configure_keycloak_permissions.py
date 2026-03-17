#!/usr/bin/env python3
"""
Script to configure Keycloak service account permissions for user management.
This grants the expense-app service account the ability to create and manage users.
"""

import urllib.request
import urllib.parse
import urllib.error
import json
import os
import time

KEYCLOAK_URL = os.environ.get("KEYCLOAK_URL", "http://localhost:8080")
ADMIN_USER = os.environ.get("KEYCLOAK_ADMIN", "admin")
ADMIN_PASSWORD = os.environ.get("KEYCLOAK_ADMIN_PASSWORD", "admin")
REALM = os.environ.get("KEYCLOAK_REALM", "expense-sharing")
CLIENT_ID = os.environ.get("KEYCLOAK_CLIENT_ID", "expense-app")

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


def get_roles_scope_id(token):
    """Get the UUID of the 'roles' client scope"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/client-scopes"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as response:
        scopes = json.loads(response.read())
        for s in scopes:
            if s["name"] == "roles":
                return s["id"]
    return None


def get_scope_mappers(token, scope_id):
    """Return list of existing mapper names for a client scope"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/client-scopes/{scope_id}/protocol-mappers/models"
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(req) as response:
            return {m["name"] for m in json.loads(response.read())}
    except urllib.error.HTTPError:
        return set()


def add_scope_mapper(token, scope_id, mapper):
    """Add a single protocol mapper to a client scope (idempotent)"""
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}/client-scopes/{scope_id}/protocol-mappers/models"
    req = urllib.request.Request(
        url,
        data=json.dumps(mapper).encode(),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        },
        method="POST"
    )
    try:
        with urllib.request.urlopen(req) as response:
            return response.status == 201
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        if e.code == 409:
            return True  # already exists
        print(f"  Warning adding mapper: {e.code} - {body}")
        return False


def fix_roles_scope_mappers(token):
    """
    Ensure the 'roles' client scope has the required protocol mappers so that
    realm roles and client roles are embedded in JWTs.
    """
    print("Checking 'roles' client scope mappers...")
    scope_id = get_roles_scope_id(token)
    if not scope_id:
        print("  WARNING: 'roles' client scope not found, skipping mapper fix.")
        return

    existing = get_scope_mappers(token, scope_id)
    print(f"  Existing mappers: {existing or 'none'}")

    mappers = [
        {
            "name": "realm roles",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-usermodel-realm-role-mapper",
            "consentRequired": False,
            "config": {
                "multivalued": "true",
                "user.attribute": "foo",
                "claim.name": "realm_access.roles",
                "jsonType.label": "String",
                "access.token.claim": "true",
                "id.token.claim": "true",
                "lightweight.claim": "false"
            }
        },
        {
            "name": "client roles",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-usermodel-client-role-mapper",
            "consentRequired": False,
            "config": {
                "multivalued": "true",
                "user.attribute": "foo",
                "claim.name": "resource_access.${client_id}.roles",
                "jsonType.label": "String",
                "access.token.claim": "true",
                "id.token.claim": "false",
                "lightweight.claim": "false"
            }
        },
        {
            "name": "audience resolve",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-audience-resolve-mapper",
            "consentRequired": False,
            "config": {
                "access.token.claim": "true",
                "id.token.claim": "false"
            }
        }
    ]

    for mapper in mappers:
        name = mapper["name"]
        if name in existing:
            print(f"  ✓ Mapper already exists: {name}")
        else:
            ok = add_scope_mapper(token, scope_id, mapper)
            print(f"  {'✓' if ok else '✗'} Added mapper: {name}")

    print("'roles' scope mappers are configured correctly.")

def wait_for_keycloak(max_retries=30, delay=5):
    """Wait for Keycloak to be reachable and ready"""
    url = f"{KEYCLOAK_URL}/health/ready"
    for attempt in range(1, max_retries + 1):
        try:
            req = urllib.request.Request(url)
            with urllib.request.urlopen(req, timeout=5) as resp:
                if resp.status == 200:
                    print(f"Keycloak is ready (attempt {attempt})")
                    return True
        except Exception as e:
            print(f"Attempt {attempt}/{max_retries}: Keycloak not ready yet ({e}), retrying in {delay}s...")
        time.sleep(delay)
    return False


def main():
    print(f"Configuring Keycloak service account permissions at {KEYCLOAK_URL}...")

    print("Waiting for Keycloak to be ready...")
    if not wait_for_keycloak():
        print("ERROR: Keycloak did not become ready in time.")
        return False

    try:
        # Get admin token
        print("Getting admin token...")
        token = get_admin_token()

        # Fix the 'roles' scope mappers FIRST so service-account tokens carry roles
        fix_roles_scope_mappers(token)
        # Refresh token after scope changes
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
