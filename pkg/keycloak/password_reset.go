package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SendPasswordResetEmail sends a password reset email to the user
func (c *Client) SendPasswordResetEmail(email string) error {
	// Get admin token
	token, err := c.getAdminToken()
	if err != nil {
		return err
	}

	// Find user by email
	userID, err := c.getUserIDByEmail(token, email)
	if err != nil {
		return err
	}

	// Send password reset email
	resetURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/execute-actions-email", c.baseURL, c.realm, userID)

	// Actions to execute (UPDATE_PASSWORD prompts user to set new password)
	actions := []string{"UPDATE_PASSWORD"}
	actionsJSON, err := json.Marshal(actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	req, err := http.NewRequest("PUT", resetURL, bytes.NewReader(actionsJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send reset email, status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("Password reset email sent", "email", email)
	return nil
}

// getUserIDByEmail finds a user ID by email address
func (c *Client) getUserIDByEmail(token, email string) (string, error) {
	searchURL := fmt.Sprintf("%s/admin/realms/%s/users?email=%s&exact=true", c.baseURL, c.realm, email)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to search user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to search user, status %d: %s", resp.StatusCode, string(body))
	}

	var users []UserRepresentation
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", fmt.Errorf("failed to decode users: %w", err)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("user not found")
	}

	return users[0].ID, nil
}
