# Testing Email Functionality with Mailhog

We have integrated **Mailhog**, a developer tool for email testing that captures all emails sent by the application.

## Accessing Mailhog

Mailhog provides a web interface where you can view all captured emails.

**Web Interface:** [http://localhost:8025](http://localhost:8025)

## How It Works

1. **Keycloak** is configured to send emails via SMTP to `mailhog:1025`.
2. **Mailhog** captures these emails instead of sending them to real addresses.
3. You can view, read, and delete these emails in the Mailhog web UI.

## Testing Password Reset

1. Go to the login page: [http://localhost:3000/login](http://localhost:3000/login)
2. Click **"Forgot password?"**
3. Enter your registered email address (e.g., `test@example.com`)
4. Open Mailhog: [http://localhost:8025](http://localhost:8025)
5. You should see a "Reset password" email from "Expense Sharing App"
6. Click on the email to view the content
7. Click the **"Link to reset password"** inside the email (or copy-paste it)
8. Set your new password

## Troubleshooting

- **No email in Mailhog?**
  - Ensure Mailhog is running: `docker compose ps mailhog`
  - Ensure Keycloak is running: `docker compose ps keycloak`
  - Verify SMTP settings in Keycloak Admin Console (Realm Settings -> Email)
    - Host: `mailhog`
    - Port: `1025`
    - From: `no-reply@expense-sharing.com`
