package utils

import (
	"net/http"
	"encoding/json"
)
////////////////////////////////////////////////////////////////////
// Assume same node as the pod
// temporary solution to get ipaddress need a better way
////////////////////////////////////////////////////////////////////

func GetPublicIPAddress() (string, error) {
	// Get the public IP address of the current node.
	resp, err := http.Get("https://api.ipify.org/?format=json")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Decode the JSON response.
	var ipAddress struct {
		IP string `json:"ip"`
	}

	err = json.NewDecoder(resp.Body).Decode(&ipAddress)
	if err != nil {
		return "", err
	}

	return ipAddress.IP, nil
}
