package paypal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAuthorization returns an authorization by ID
// Endpoint: GET /v2/payments/authorization/ID
func (c *Client) GetAuthorization(authID string) (*Authorization, error) {
	buf := bytes.NewBuffer([]byte(""))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s%s", c.APIBase, "/v2/payments/authorizations/", authID), buf)
	auth := &Authorization{}

	if err != nil {
		return auth, err
	}

	err = c.SendWithAuth(req, auth)
	return auth, err
}

// CaptureAuthorization captures and process an existing authorization.
// To use this method, the original payment must have Intent set to "authorize"
// Endpoint: POST /v2/payments/authorizations/ID/capture
func (c *Client) CaptureAuthorization(authID string, paymentCaptureRequest *PaymentCaptureRequest) (*PaymentCaptureResponse, error) {
	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/payments/authorizations/"+authID+"/capture"), paymentCaptureRequest)
	paymentCaptureResponse := &PaymentCaptureResponse{}

	req.Header.Set(HeaderPrefer, HeaderPreferRepresentation)

	if err != nil {
		return paymentCaptureResponse, err
	}

	err = c.SendWithAuth(req, paymentCaptureResponse)
	return paymentCaptureResponse, err
}

// VoidAuthorization voids a previously authorized payment
// Endpoint: POST /v2/payments/authorization/ID/void
func (c *Client) VoidAuthorization(authID string) error {

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/payments/authorizations/"+authID+"/void"), nil)

	if err != nil {
		return err
	}

	err = c.SendWithAuth(req, nil)
	return err
}

// ReauthorizeAuthorization reauthorize a Paypal account payment.
// PayPal recommends to reauthorize payment after ~3 days
// Endpoint: POST /v2/payments/authorization/ID/reauthorize
func (c *Client) ReauthorizeAuthorization(authID string, a *Amount) (*Authorization, error) {
	// buf := bytes.NewBuffer([]byte(`{"amount":{"currenct_":"` + a.Currency + `","total":"` + a.Total + `"}}`))
	body, berr := json.Marshal(&struct {
		Amount *Amount `json:"amount,omitempty"`
	}{a})
	if berr != nil {
		return nil, berr
	}
	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.APIBase, "/v2/payments/authorizations/"+authID+"/reauthorize"), buf)
	auth := &Authorization{}

	if err != nil {
		return auth, err
	}

	err = c.SendWithAuth(req, auth)
	return auth, err
}
