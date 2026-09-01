package queryx

import "sort"

// SupportedTypes returns the sorted list of server type identifiers that are
// registered on this client and can be passed to Query. The result reflects the
// protocols actually registered (see RegisterDefaultProtocols), so a client
// created with NewClient (no protocols) returns an empty slice while one created
// with NewClientWithDefaults returns every built-in type.
func (c *Client) SupportedTypes() []string {
	types := c.factory.List()
	sort.Strings(types)
	return types
}

// Supports reports whether the given server type is registered on this client
// and therefore queryable. It is a cheap, allocation-free membership check that
// callers can use to validate user input before issuing a Query.
func (c *Client) Supports(serverType ServerType) bool {
	_, err := c.factory.Get(string(serverType))
	return err == nil
}
