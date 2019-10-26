package paypal

import "fmt"

// GetOrder retrieves order by ID
// Endpoint: GET /v2/checkout/orders/ID
func (c *Client) GetOrder(orderID string) (*Order, error) {
	order := &Order{}

	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s%s", c.APIBase, "/v2/checkout/orders/", orderID), nil)
	if err != nil {
		return order, err
	}

	if err = c.SendWithAuth(req, order); err != nil {
		return order, err
	}

	return order, nil
}

// CreateOrder - Use this call to create an order
// Endpoint: POST /v2/checkout/orders
func (c *Client) CreateOrder(intent PaymentIntent, purchaseUnits []PurchaseUnitRequest, payer *CreateOrderPayer, appContext *ApplicationContext) (*Order, error) {
	type createOrderRequest struct {
		Intent             PaymentIntent         `json:"intent"`
		Payer              *CreateOrderPayer     `json:"payer,omitempty"`
		PurchaseUnits      []PurchaseUnitRequest `json:"purchase_units"`
		ApplicationContext *ApplicationContext   `json:"application_context,omitempty"`
	}

	order := &Order{}

	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/checkout/orders"), createOrderRequest{Intent: intent, PurchaseUnits: purchaseUnits, Payer: payer, ApplicationContext: appContext})
	if err != nil {
		return order, err
	}

	if err = c.SendWithAuth(req, order); err != nil {
		return order, err
	}

	return order, nil
}

// UpdateOrder updates the order by ID
// Endpoint: PATCH /v2/checkout/orders/ID
func (c *Client) UpdateOrder(orderID string, orderUpdate []PaymentPatch) error {

	req, err := c.NewRequest("PATCH", fmt.Sprintf("%s%s%s", c.APIBase, "/v2/checkout/orders/", orderID), orderUpdate)
	if err != nil {
		return err
	}

	if err = c.SendWithAuth(req, nil); err != nil {
		return err
	}

	return nil
}

// AuthorizeOrder - https://developer.paypal.com/docs/api/orders/v2/#orders_authorize
// Endpoint: POST /v2/checkout/orders/ID/authorize
func (c *Client) AuthorizeOrder(orderID string, authorizeOrderRequest AuthorizeOrderRequest) (*AuthorizeOrderResponse, error) {
	auth := &AuthorizeOrderResponse{}

	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/checkout/orders/"+orderID+"/authorize"), authorizeOrderRequest)
	if err != nil {
		return auth, err
	}

	if err = c.SendWithAuth(req, auth); err != nil {
		return auth, err
	}

	return auth, nil
}

// CaptureOrder - https://developer.paypal.com/docs/api/orders/v2/#orders_capture
// Endpoint: POST /v2/checkout/orders/ID/capture
func (c *Client) CaptureOrder(orderID string, captureOrderRequest CaptureOrderRequest) (*CaptureOrderResponse, error) {
	capture := &CaptureOrderResponse{}

	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/checkout/orders/"+orderID+"/capture"), captureOrderRequest)
	if err != nil {
		return capture, err
	}

	if err = c.SendWithAuth(req, capture); err != nil {
		return capture, err
	}

	return capture, nil
}

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
func (c *Client) RefundPayment(captureID string, request RefundRequest) (*RefundResponse, error) {
	refund := new(RefundResponse)

	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/v2/payments/captures/"+captureID+"/refund"), request)
	if err != nil {
		return nil, err
	}

	if err = c.SendWithAuth(req, refund); err != nil {
		return nil, err
	}

	return refund, nil
}
