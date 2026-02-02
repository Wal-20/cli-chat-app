package client

import "fmt"

// Admin actions
func (c *APIClient) InviteUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/invite/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) KickUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/kick/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) BanUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/ban/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) MakeAdmin(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/promote/%s", chatroomID, userID), nil)
	return err
}
