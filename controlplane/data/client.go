package data

type ClientUID string

type Client struct {
	UID ClientUID
}

func NewClient(uid ClientUID) *Client {
	return &Client{
		UID: uid,
	}
}
