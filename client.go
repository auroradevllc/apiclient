package apiclient

var DefaultClient = &Client{}

type Client struct {
	defaultOpts []Option
}

func NewClient(opts ...Option) *Client {
	return &Client{
		defaultOpts: opts,
	}
}
