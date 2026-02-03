import json
import urllib.request
import urllib.parse
import ssl

# Configuration
KEYCLOAK_URL = "http://localhost:8080"
REALM = "expense-sharing"
ADMIN_USER = "admin"
ADMIN_PASS = "admin"

def get_admin_token():
    url = f"{KEYCLOAK_URL}/realms/master/protocol/openid-connect/token"
    data = urllib.parse.urlencode({
        "client_id": "admin-cli",
        "username": ADMIN_USER,
        "password": ADMIN_PASS,
        "grant_type": "password"
    }).encode()
    
    req = urllib.request.Request(url, data=data, method="POST")
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())["access_token"]

def get_realm(token):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}"
    req = urllib.request.Request(url, method="GET")
    req.add_header("Authorization", f"Bearer {token}")
    
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())

def update_realm(token, realm_data):
    url = f"{KEYCLOAK_URL}/admin/realms/{REALM}"
    data = json.dumps(realm_data).encode()
    
    req = urllib.request.Request(url, data=data, method="PUT")
    req.add_header("Authorization", f"Bearer {token}")
    req.add_header("Content-Type", "application/json")
    
    try:
        with urllib.request.urlopen(req) as response:
            print("Realm updated successfully!")
    except urllib.error.HTTPError as e:
        print(f"Failed to update realm: {e.code}")
        print(e.read().decode())
        exit(1)

def main():
    print("Getting admin token...")
    token = get_admin_token()
    
    print(f"Fetching realm '{REALM}'...")
    realm = get_realm(token)
    
    print("Configuring SMTP...")
    realm["smtpServer"] = {
        "replyToDisplayName": "Expense Sharing Support",
        "starttls": "false",
        "auth": "",
        "port": "1025",
        "host": "mailhog",
        "replyTo": "support@expense-sharing.com",
        "from": "no-reply@expense-sharing.com",
        "fromDisplayName": "Expense Sharing App",
        "ssl": "false"
    }
    
    print("Updating realm...")
    update_realm(token, realm)
    print("Done! SMTP configured to use Mailhog.")

if __name__ == "__main__":
    main()
