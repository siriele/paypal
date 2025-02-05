package paypal

import (
	"fmt"
	"net/http"
)

/*

curl -v -X POST https://api.sandbox.paypal.com/v2/payments/captures/2GG279541U471931P/refund \
-H "Content-Type: application/json" \
-H "Authorization: Bearer Access-Token" \
-H "PayPal-Request-Id: 123e4567-e89b-12d3-a456-426655440020" \
-d '{
  "amount": {
    "value": "10.99",
    "currency_code": "USD"
  },
  "invoice_id": "INVOICE-123",
  "note_to_payer": "Defective product"
}'
*/

// RefundCapture - https://developer.paypal.com/docs/api/payments/v2/#captures_refund
// Endpoint: POST /v2/payments/captures/ID/refund
func (c *Client) RefundCapture(captureID string, request *RefundRequest) (*RefundResponse, error) {
	refund := new(RefundResponse)

	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/payments/captures/"+captureID+"/refund"), request)
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderPrefer, HeaderPreferRepresentation)
	if err = c.SendWithAuth(req, refund); err != nil {
		return nil, err
	}

	return refund, nil
}

func (c *Client) UpdateTracking(request *TrackersRequest) (*TrackersResponse, error) {
	response := new(TrackersResponse)

	req, err := c.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.APIBase, "/v1/shipping/trackers-batch"), request)
	if err != nil {
		return nil, err
	}
	if err = c.SendWithAuth(req, response); err != nil {
		return nil, err
	}

	return response, nil
}
