package paypal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// APIBaseSandBox points to the sandbox (for testing) version of the API
	APIBaseSandBox = "https://api.sandbox.paypal.com"

	// APIBaseLive points to the live version of the API
	APIBaseLive = "https://api.paypal.com"

	// RequestNewTokenBeforeExpiresIn is used by SendWithAuth and try to get new Token when it's about to expire
	RequestNewTokenBeforeExpiresIn = time.Duration(60) * time.Second
)

// Possible values for `no_shipping` in InputFields
//
// https://developer.paypal.com/docs/api/payment-experience/#definition-input_fields
const (
	NoShippingDisplay      uint = 0
	NoShippingHide         uint = 1
	NoShippingBuyerAccount uint = 2
)

// Possible values for `address_override` in InputFields
//
// https://developer.paypal.com/docs/api/payment-experience/#definition-input_fields
const (
	AddrOverrideFromFile uint = 0
	AddrOverrideFromCall uint = 1
)

// Possible values for `landing_page_type` in FlowConfig
//
// https://developer.paypal.com/docs/api/payment-experience/#definition-flow_config
const (
	LandingPageTypeBilling string = "Billing"
	LandingPageTypeLogin   string = "Login"
)

// Possible value for `allowed_payment_method` in PaymentOptions
//
// https://developer.paypal.com/docs/api/payments/#definition-payment_options
const (
	AllowedPaymentUnrestricted         string = "UNRESTRICTED"
	AllowedPaymentInstantFundingSource string = "INSTANT_FUNDING_SOURCE"
	AllowedPaymentImmediatePay         string = "IMMEDIATE_PAY"
)

// Possible value for `intent` in CreateOrder
//
// https://developer.paypal.com/docs/api/orders/v2/#orders_create
const (
	OrderIntentCapture   string = "CAPTURE"
	OrderIntentAuthorize string = "AUTHORIZE"
)

type ItemCategory string

// Possible values for `category` in Item
//
// https://developer.paypal.com/docs/api/orders/v2/#definition-item
const (
	ItemCategoryDigitalGood  ItemCategory = "DIGITAL_GOODS"
	ItemCategoryPhysicalGood ItemCategory = "PHYSICAL_GOODS"
)

// Possible values for `shipping_preference` in ApplicationContext
//
// https://developer.paypal.com/docs/api/orders/v2/#definition-application_context
const (
	ShippingPreferenceGetFromFile        string = "GET_FROM_FILE"
	ShippingPreferenceNoShipping         string = "NO_SHIPPING"
	ShippingPreferenceSetProvidedAddress string = "SET_PROVIDED_ADDRESS"
)

const (
	EventPaymentCaptureCompleted       string = "PAYMENT.CAPTURE.COMPLETED"
	EventPaymentCaptureDenied          string = "PAYMENT.CAPTURE.DENIED"
	EventPaymentCaptureRefunded        string = "PAYMENT.CAPTURE.REFUNDED"
	EventMerchantOnboardingCompleted   string = "MERCHANT.ONBOARDING.COMPLETED"
	EventMerchantPartnerConsentRevoked string = "MERCHANT.PARTNER-CONSENT.REVOKED"
)

const (
	OperationAPIIntegration   string = "API_INTEGRATION"
	ProductExpressCheckout    string = "EXPRESS_CHECKOUT"
	IntegrationMethodPayPal   string = "PAYPAL"
	IntegrationTypeThirdParty string = "THIRD_PARTY"
	ConsentShareData          string = "SHARE_DATA_CONSENT"
)

const (
	FeaturePayment               string = "PAYMENT"
	FeatureRefund                string = "REFUND"
	FeatureFuturePayment         string = "FUTURE_PAYMENT"
	FeatureDirectPayment         string = "DIRECT_PAYMENT"
	FeaturePartnerFee            string = "PARTNER_FEE"
	FeatureDelayFunds            string = "DELAY_FUNDS_DISBURSEMENT"
	FeatureReadSellerDispute     string = "READ_SELLER_DISPUTE"
	FeatureUpdateSellerDispute   string = "UPDATE_SELLER_DISPUTE"
	FeatureDisputeReadBuyer      string = "DISPUTE_READ_BUYER"
	FeatureUpdateCustomerDispute string = "UPDATE_CUSTOMER_DISPUTES"
)

const (
	LinkRelSelf      string = "self"
	LinkRelActionURL string = "action_url"
)

type (
	// JSONTime overrides MarshalJson method to format in ISO8601
	JSONTime time.Time

	// Address struct
	Address struct {
		Line1       string `json:"line1"`
		Line2       string `json:"line2,omitempty"`
		City        string `json:"city"`
		CountryCode string `json:"country_code"`
		PostalCode  string `json:"postal_code,omitempty"`
		State       string `json:"state,omitempty"`
		Phone       string `json:"phone,omitempty"`
	}

	// AgreementDetails struct
	AgreementDetails struct {
		OutstandingBalance AmountPayout `json:"outstanding_balance"`
		CyclesRemaining    int          `json:"cycles_remaining,string"`
		CyclesCompleted    int          `json:"cycles_completed,string"`
		NextBillingDate    time.Time    `json:"next_billing_date"`
		LastPaymentDate    time.Time    `json:"last_payment_date"`
		LastPaymentAmount  AmountPayout `json:"last_payment_amount"`
		FinalPaymentDate   time.Time    `json:"final_payment_date"`
		FailedPaymentCount int          `json:"failed_payment_count,string"`
	}

	// Amount struct
	Amount struct {
		Currency  string           `json:"currency_code"`
		Value     string           `json:"value"`
		Breakdown *AmountBreakdown `json:"breakdown,omitempty"`
	}

	// AmountPayout struct
	AmountPayout struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	}

	// ApplicationContext struct
	ApplicationContext struct {
		BrandName          string `json:"brand_name,omitempty"`
		Locale             string `json:"locale,omitempty"`
		LandingPage        string `json:"landing_page,omitempty"`
		ShippingPreference string `json:"shipping_preference,omitempty"`
		UserAction         string `json:"user_action,omitempty"`
		ReturnURL          string `json:"return_url,omitempty"`
		CancelURL          string `json:"cancel_url,omitempty"`
	}

	// Authorization struct
	Authorization struct {
		ID               string                `json:"id,omitempty"`
		CustomID         string                `json:"custom_id,omitempty"`
		InvoiceID        string                `json:"invoice_id,omitempty"`
		Status           AuthorizationStatus   `json:"status,omitempty"`
		StatusDetails    *CaptureStatusDetails `json:"status_details,omitempty"`
		Amount           *PurchaseUnitAmount   `json:"amount,omitempty"`
		SellerProtection *SellerProtection     `json:"seller_protection,omitempty"`
		CreateTime       PTime                 `json:"create_time,omitempty"`
		UpdateTime       PTime                 `json:"update_time,omitempty"`
		ExpirationTime   PTime                 `json:"expiration_time,omitempty"`
		Links            []Link                `json:"links,omitempty"`
	}

	// AuthorizeOrderResponse .
	AuthorizeOrderResponse struct {
		CreateTime    PTime                  `json:"create_time,omitempty"`
		UpdateTime    PTime                  `json:"update_time,omitempty"`
		ID            string                 `json:"id,omitempty"`
		Status        OrderStatus            `json:"status,omitempty"`
		Intent        PaymentIntent          `json:"intent,omitempty"`
		PurchaseUnits []PurchaseUnit         `json:"purchase_units,omitempty"`
		Payer         *PayerWithNameAndPhone `json:"payer,omitempty"`
	}

	// AuthorizeOrderRequest - https://developer.paypal.com/docs/api/orders/v2/#orders_authorize
	AuthorizeOrderRequest struct {
		PaymentSource      *PaymentSource     `json:"payment_source,omitempty"`
		ApplicationContext ApplicationContext `json:"application_context,omitempty"`
	}

	// https://developer.paypal.com/docs/api/payments/v2/#definition-platform_fee
	PlatformFee struct {
		Amount *Money          `json:"amount,omitempty"`
		Payee  *PayeeForOrders `json:"payee,omitempty"`
	}

	// https://developer.paypal.com/docs/api/payments/v2/#definition-payment_instruction
	PaymentInstruction struct {
		PlatformFees     []PlatformFee `json:"platform_fees,omitempty"`
		DisbursementMode string        `json:"disbursement_mode,omitempty"`
	}

	// https://developer.paypal.com/docs/api/payments/v2/#authorizations_capture
	PaymentCaptureRequest struct {
		InvoiceID           string             `json:"invoice_id,omitempty"`
		NoteToPayer         string             `json:"note_to_payer,omitempty"`
		SoftDescriptor      string             `json:"soft_descriptor,omitempty"`
		Amount              *Money             `json:"amount,omitempty"`
		FinalCapture        bool               `json:"final_capture,omitempty"`
		PaymentInstructions PaymentInstruction `json:"payment_instruction"`
	}

	SellerProtection struct {
		Status            string   `json:"status,omitempty"`
		DisputeCategories []string `json:"dispute_categories,omitempty"`
	}

	// https://developer.paypal.com/docs/api/payments/v2/#definition-capture_status_details
	CaptureStatusDetails struct {
		Reason string `json:"reason,omitempty"`
	}

	PaymentCaptureResponse struct {
		ID               string                     `json:"id,omitempty"`
		Status           CaptureStatus              `json:"status,omitempty"`
		StatusDetails    *CaptureStatusDetails      `json:"status_details,omitempty"`
		Amount           *Money                     `json:"amount,omitempty"`
		InvoiceID        string                     `json:"invoice_id,omitempty"`
		CustomID         string                     `json:"custom_id,omitempty"`
		FinalCapture     bool                       `json:"final_capture,omitempty"`
		DisbursementMode string                     `json:"disbursement_mode,omitempty"`
		Breakdown        *SellerReceivableBreakdown `json:"seller_receivable_breakdown,omitempty"`
		CreateTime       PTime                      `json:"create_time,omitempty"`
		UpdateTime       PTime                      `json:"update_time,omitempty"`
	}

	// CaptureOrderRequest - https://developer.paypal.com/docs/api/orders/v2/#orders_capture
	CaptureOrderRequest struct {
		PaymentSource *PaymentSource `json:"payment_source"`
	}

	// BatchHeader struct
	BatchHeader struct {
		Amount            *AmountPayout      `json:"amount,omitempty"`
		Fees              *AmountPayout      `json:"fees,omitempty"`
		PayoutBatchID     string             `json:"payout_batch_id,omitempty"`
		BatchStatus       string             `json:"batch_status,omitempty"`
		TimeCreated       PTime              `json:"time_created,omitempty"`
		TimeCompleted     PTime              `json:"time_completed,omitempty"`
		SenderBatchHeader *SenderBatchHeader `json:"sender_batch_header,omitempty"`
	}

	// BillingAgreement struct
	BillingAgreement struct {
		Name                        string               `json:"name,omitempty"`
		Description                 string               `json:"description,omitempty"`
		StartDate                   JSONTime             `json:"start_date,omitempty"`
		Plan                        BillingPlan          `json:"plan,omitempty"`
		Payer                       Payer                `json:"payer,omitempty"`
		ShippingAddress             *ShippingAddress     `json:"shipping_address,omitempty"`
		OverrideMerchantPreferences *MerchantPreferences `json:"override_merchant_preferences,omitempty"`
	}

	// BillingPlan struct
	BillingPlan struct {
		ID                  string               `json:"id,omitempty"`
		Name                string               `json:"name,omitempty"`
		Description         string               `json:"description,omitempty"`
		Type                string               `json:"type,omitempty"`
		PaymentDefinitions  []PaymentDefinition  `json:"payment_definitions,omitempty"`
		MerchantPreferences *MerchantPreferences `json:"merchant_preferences,omitempty"`
	}

	// Capture struct
	Capture struct {
		Amount         *Amount                    `json:"amount,omitempty"`
		IsFinalCapture bool                       `json:"final_capture"`
		CreateTime     PTime                      `json:"create_time,omitempty"`
		UpdateTime     PTime                      `json:"update_time,omitempty"`
		Status         CaptureStatus              `json:"status,omitempty"`
		ID             string                     `json:"id,omitempty"`
		InvoiceID      string                     `json:"invoice_id,omitempty"`
		Links          []Link                     `json:"links,omitempty"`
		Breakdown      *SellerReceivableBreakdown `json:"seller_receivable_breakdown,omitempty"`
	}

	// ChargeModel struct
	ChargeModel struct {
		Type   string       `json:"type,omitempty"`
		Amount AmountPayout `json:"amount,omitempty"`
	}

	// Client represents a Paypal REST API Client
	Client struct {
		sync.Mutex
		Client         *http.Client
		ClientID       string
		Secret         string
		APIBase        string
		Log            io.Writer // If user set log file name all requests will be logged there
		Token          *TokenResponse
		tokenExpiresAt time.Time
	}

	// CreditCard struct
	CreditCard struct {
		ID                 string   `json:"id,omitempty"`
		PayerID            string   `json:"payer_id,omitempty"`
		ExternalCustomerID string   `json:"external_customer_id,omitempty"`
		Number             string   `json:"number"`
		Type               string   `json:"type"`
		ExpireMonth        string   `json:"expire_month"`
		ExpireYear         string   `json:"expire_year"`
		CVV2               string   `json:"cvv2,omitempty"`
		FirstName          string   `json:"first_name,omitempty"`
		LastName           string   `json:"last_name,omitempty"`
		BillingAddress     *Address `json:"billing_address,omitempty"`
		State              string   `json:"state,omitempty"`
		ValidUntil         string   `json:"valid_until,omitempty"`
	}

	// CreditCards GET /v1/vault/credit-cards
	CreditCards struct {
		Items      []CreditCard `json:"items"`
		Links      []Link       `json:"links"`
		TotalItems int          `json:"total_items"`
		TotalPages int          `json:"total_pages"`
	}

	// CreditCardToken struct
	CreditCardToken struct {
		CreditCardID string `json:"credit_card_id"`
		PayerID      string `json:"payer_id,omitempty"`
		Last4        string `json:"last4,omitempty"`
		ExpireYear   string `json:"expire_year,omitempty"`
		ExpireMonth  string `json:"expire_month,omitempty"`
	}

	// CreditCardsFilter struct
	CreditCardsFilter struct {
		PageSize int
		Page     int
	}

	// CreditCardField PATCH /v1/vault/credit-cards/credit_card_id
	CreditCardField struct {
		Operation string `json:"op"`
		Path      string `json:"path"`
		Value     string `json:"value"`
	}

	// Currency struct
	Currency struct {
		Currency string `json:"currency,omitempty"`
		Value    string `json:"value,omitempty"`
	}

	// Details structure used in Amount structures as optional value
	Details struct {
		Subtotal         *Money `json:"subtotal,omitempty"`
		Shipping         *Money `json:"shipping,omitempty"`
		Tax              *Money `json:"tax,omitempty"`
		HandlingFee      *Money `json:"handling_fee,omitempty"`
		ShippingDiscount *Money `json:"shipping_discount,omitempty"`
		Insurance        *Money `json:"insurance,omitempty"`
		GiftWrap         *Money `json:"gift_wrap,omitempty"`
	}

	// ErrorResponseDetail struct
	ErrorResponseDetail struct {
		Field string `json:"field"`
		Issue string `json:"issue"`
		Links []Link `json:"link"`
	}

	// ErrorResponse https://developer.paypal.com/docs/api/errors/
	ErrorResponse struct {
		Response        *http.Response        `json:"-"`
		Name            string                `json:"name"`
		DebugID         string                `json:"debug_id"`
		Message         string                `json:"message"`
		InformationLink string                `json:"information_link"`
		Details         []ErrorResponseDetail `json:"details"`
	}

	// ExecuteAgreementResponse struct
	ExecuteAgreementResponse struct {
		ID               string           `json:"id"`
		State            string           `json:"state"`
		Description      string           `json:"description,omitempty"`
		Payer            Payer            `json:"payer"`
		Plan             BillingPlan      `json:"plan"`
		StartDate        time.Time        `json:"start_date"`
		ShippingAddress  ShippingAddress  `json:"shipping_address"`
		AgreementDetails AgreementDetails `json:"agreement_details"`
		Links            []Link           `json:"links"`
	}

	// ExecuteResponse struct
	ExecuteResponse struct {
		ID           string        `json:"id"`
		Links        []Link        `json:"links"`
		State        string        `json:"state"`
		Payer        PaymentPayer  `json:"payer"`
		Transactions []Transaction `json:"transactions,omitempty"`
	}

	// FundingInstrument struct
	FundingInstrument struct {
		CreditCard      *CreditCard      `json:"credit_card,omitempty"`
		CreditCardToken *CreditCardToken `json:"credit_card_token,omitempty"`
	}

	// Item struct
	Item struct {
		Quantity    string       `json:"quantity"`
		Name        string       `json:"name"`
		Price       string       `json:"price"`
		Currency    string       `json:"currency"`
		SKU         string       `json:"sku,omitempty"`
		Description string       `json:"description,omitempty"`
		Tax         *Money       `json:"tax,omitempty"`
		UnitAmount  *Money       `json:"unit_amount,omitempty"`
		Category    ItemCategory `json:"category,omitempty"`
	}

	// ItemList struct
	ItemList struct {
		Items           []Item           `json:"items,omitempty"`
		ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"`
	}

	// Link struct
	Link struct {
		Href        string `json:"href"`
		Rel         string `json:"rel,omitempty"`
		Method      string `json:"method,omitempty"`
		Description string `json:"description,omitempty"`
		Enctype     string `json:"enctype,omitempty"`
	}

	// PurchaseUnitAmount struct
	PurchaseUnitAmount struct {
		Currency  string                       `json:"currency_code"`
		Value     string                       `json:"value"`
		Breakdown *PurchaseUnitAmountBreakdown `json:"breakdown,omitempty"`
	}
	ExchangeRate struct {
		SourceCurrency string `json:"source_currency"`
		TargetCurrency string `json:"target_currency"`
		Value          string `json:"value"`
	}

	SellerReceivableBreakdown struct {
		// The amount for this captured payment.
		GrossAmount *Money `json:"gross_amount,omitempty"`

		// paypal_fee object
		// The applicable fee for this captured payment.
		// Read only.
		PaypalFee *Money `json:"paypal_fee,omitempty"`
		// net_amount object
		// The net amount that the payee receives for this captured payment in their PayPal account. The net amount is computed as gross_amount minus the paypal_fee minus the platform_fees.
		// Read only.
		NetAmount *Money `json:"net_amount,omitempty"`
		// receivable_amount object
		// The net amount that is credited to the payee's PayPal account. Returned only when the currency of the captured payment is different from the currency of the PayPal account where the payee wants to credit the funds. The amount is computed as net_amount times exchange_rate.
		// Read only.
		ReceivableAmount *Money `json:"receivable_amount,omitempty"`
		// exchange_rate object
		// The exchange rate that determines the amount that is credited to the payee's PayPal account. Returned when the currency of the captured payment is different from the currency of the PayPal account where the payee wants to credit the funds.
		// Read only.
		ExchangeRate *ExchangeRate `json:"exchange_rate,omitempty"`
		// platform_fees array (contains the platform_fee object)
		// An array of platform or partner fees, commissions, or brokerage fees that associated with the captured payment.
		// Read only.
		PlatformFees []PlatformFee `json:"platform_fees,omitempty"`
	}
	// PurchaseUnitAmountBreakdown struct
	PurchaseUnitAmountBreakdown struct {
		ItemTotal        *Money `json:"item_total,omitempty"`
		Shipping         *Money `json:"shipping,omitempty"`
		Handling         *Money `json:"handling,omitempty"`
		TaxTotal         *Money `json:"tax_total,omitempty"`
		Insurance        *Money `json:"insurance,omitempty"`
		ShippingDiscount *Money `json:"shipping_discount,omitempty"`
		Discount         *Money `json:"discount,omitempty"`
	}
	AmountBreakdown struct {
		ItemTotal        *Money `json:"item_total,omitempty"`
		Shipping         *Money `json:"shipping,omitempty"`
		Handling         *Money `json:"handling,omitempty"`
		TaxTotal         *Money `json:"tax_total,omitempty"`
		Insurance        *Money `json:"insurance,omitempty"`
		ShippingDiscount *Money `json:"shipping_discount,omitempty"`
		Discount         *Money `json:"discount,omitempty"`
	}
	// Money struct
	//
	// https://developer.paypal.com/docs/api/orders/v2/#definition-money
	Money struct {
		Currency string `json:"currency_code"`
		Value    string `json:"value"`
	}

	PurchaseUnitPayments struct {
		Authorizations []Authorization `json:"authorizations,omitempty"`
		Captures       []Capture       `json:"captures,omitempty"`
		Refunds        []Refund        `json:"refunds,omitempty"`
	}
	// PurchaseUnit struct
	PurchaseUnit struct {
		ReferenceID    string                `json:"reference_id"`
		Amount         *PurchaseUnitAmount   `json:"amount,omitempty"`
		Payee          *PayeeForOrders       `json:"payee,omitempty"`
		Description    string                `json:"description,omitempty"`
		CustomID       string                `json:"custom_id,omitempty"`
		InvoiceID      string                `json:"invoice_id,omitempty"`
		SoftDescriptor string                `json:"soft_descriptor,omitempty"`
		Items          []Item                `json:"items,omitempty"`
		Shipping       *ShippingDetail       `json:"shipping,omitempty"`
		Payments       *PurchaseUnitPayments `json:"payments,omitempty"`
	}

	// TaxInfo used for orders.
	TaxInfo struct {
		TaxID     string `json:"tax_id,omitempty"`
		TaxIDType string `json:"tax_id_type,omitempty"`
	}

	// PhoneWithTypeNumber struct for PhoneWithType
	PhoneWithTypeNumber struct {
		NationalNumber string `json:"national_number,omitempty"`
	}

	// PhoneWithType struct used for orders
	PhoneWithType struct {
		PhoneType   string               `json:"phone_type,omitempty"`
		PhoneNumber *PhoneWithTypeNumber `json:"phone_number,omitempty"`
	}

	// CreateOrderPayerName create order payer name
	CreateOrderPayerName struct {
		GivenName string `json:"given_name,omitempty"`
		Surname   string `json:"surname,omitempty"`
	}

	// CreateOrderPayer used with create order requests
	CreateOrderPayer struct {
		Name         *CreateOrderPayerName          `json:"name,omitempty"`
		EmailAddress string                         `json:"email_address,omitempty"`
		PayerID      string                         `json:"payer_id,omitempty"`
		Phone        *PhoneWithType                 `json:"phone,omitempty"`
		BirthDate    string                         `json:"birth_date,omitempty"`
		TaxInfo      *TaxInfo                       `json:"tax_info,omitempty"`
		Address      *ShippingDetailAddressPortable `json:"address,omitempty"`
	}

	// PurchaseUnitRequest struct
	PurchaseUnitRequest struct {
		ReferenceID    string              `json:"reference_id,omitempty"`
		Amount         *PurchaseUnitAmount `json:"amount"`
		Payee          *PayeeForOrders     `json:"payee,omitempty"`
		Description    string              `json:"description,omitempty"`
		CustomID       string              `json:"custom_id,omitempty"`
		InvoiceID      string              `json:"invoice_id,omitempty"`
		SoftDescriptor string              `json:"soft_descriptor,omitempty"`
		Items          []Item              `json:"items,omitempty"`
		Shipping       *ShippingDetail     `json:"shipping,omitempty"`
	}

	// MerchantPreferences struct
	MerchantPreferences struct {
		SetupFee                *AmountPayout `json:"setup_fee,omitempty"`
		ReturnURL               string        `json:"return_url,omitempty"`
		CancelURL               string        `json:"cancel_url,omitempty"`
		AutoBillAmount          string        `json:"auto_bill_amount,omitempty"`
		InitialFailAmountAction string        `json:"initial_fail_amount_action,omitempty"`
		MaxFailAttempts         string        `json:"max_fail_attempts,omitempty"`
	}

	// Order struct
	Order struct {
		ID            string         `json:"id,omitempty"`
		Status        OrderStatus    `json:"status,omitempty"`
		Intent        PaymentIntent  `json:"intent,omitempty"`
		PurchaseUnits []PurchaseUnit `json:"purchase_units,omitempty"`
		Links         []Link         `json:"links,omitempty"`
		CreateTime    PTime          `json:"create_time,omitempty"`
		UpdateTime    PTime          `json:"update_time,omitempty"`
	}

	// CaptureAmount struct
	CaptureAmount struct {
		ID       string              `json:"id,omitempty"`
		CustomID string              `json:"custom_id,omitempty"`
		Amount   *PurchaseUnitAmount `json:"amount,omitempty"`
	}

	// CapturedPayments has the amounts for a captured order
	CapturedPayments struct {
		Captures []CaptureAmount `json:"captures,omitempty"`
	}

	// CapturedPurchaseUnit are purchase units for a captured order
	CapturedPurchaseUnit struct {
		Payments *CapturedPayments `json:"payments,omitempty"`
	}

	// PayerWithNameAndPhone struct
	PayerWithNameAndPhone struct {
		Name         *CreateOrderPayerName `json:"name,omitempty"`
		EmailAddress string                `json:"email_address,omitempty"`
		Phone        *PhoneWithType        `json:"phone,omitempty"`
		PayerID      string                `json:"payer_id,omitempty"`
	}

	// CaptureOrderResponse is the response for capture order
	CaptureOrderResponse struct {
		ID            string                 `json:"id,omitempty"`
		Status        OrderStatus            `json:"status,omitempty"`
		Payer         *PayerWithNameAndPhone `json:"payer,omitempty"`
		PurchaseUnits []PurchaseUnit         `json:"purchase_units,omitempty"`
	}

	// Payer struct
	Payer struct {
		PaymentMethod      string              `json:"payment_method"`
		FundingInstruments []FundingInstrument `json:"funding_instruments,omitempty"`
		PayerInfo          *PayerInfo          `json:"payer_info,omitempty"`
		Status             string              `json:"payer_status,omitempty"`
	}

	// PayerInfo struct
	PayerInfo struct {
		Email           string           `json:"email,omitempty"`
		FirstName       string           `json:"first_name,omitempty"`
		LastName        string           `json:"last_name,omitempty"`
		PayerID         string           `json:"payer_id,omitempty"`
		Phone           string           `json:"phone,omitempty"`
		ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"`
		TaxIDType       string           `json:"tax_id_type,omitempty"`
		TaxID           string           `json:"tax_id,omitempty"`
		CountryCode     string           `json:"country_code"`
	}

	// PaymentDefinition struct
	PaymentDefinition struct {
		ID                string        `json:"id,omitempty"`
		Name              string        `json:"name,omitempty"`
		Type              string        `json:"type,omitempty"`
		Frequency         string        `json:"frequency,omitempty"`
		FrequencyInterval string        `json:"frequency_interval,omitempty"`
		Amount            AmountPayout  `json:"amount,omitempty"`
		Cycles            string        `json:"cycles,omitempty"`
		ChargeModels      []ChargeModel `json:"charge_models,omitempty"`
	}

	// PaymentOptions struct
	PaymentOptions struct {
		AllowedPaymentMethod string `json:"allowed_payment_method,omitempty"`
	}

	// PaymentPatch PATCH /v2/payments/payment/{payment_id)
	PaymentPatch struct {
		Operation string      `json:"op"`
		Path      string      `json:"path"`
		Value     interface{} `json:"value"`
	}

	// PaymentPayer struct
	PaymentPayer struct {
		PaymentMethod string     `json:"payment_method"`
		Status        string     `json:"status,omitempty"`
		PayerInfo     *PayerInfo `json:"payer_info,omitempty"`
	}

	// PaymentResponse structure
	PaymentResponse struct {
		ID           string        `json:"id"`
		State        string        `json:"state"`
		Intent       string        `json:"intent"`
		Payer        Payer         `json:"payer"`
		Transactions []Transaction `json:"transactions"`
		Links        []Link        `json:"links"`
	}

	// PaymentSource structure
	PaymentSource struct {
		Card  *PaymentSourceCard  `json:"card"`
		Token *PaymentSourceToken `json:"token"`
	}

	// PaymentSourceCard structure
	PaymentSourceCard struct {
		ID             string              `json:"id"`
		Name           string              `json:"name"`
		Number         string              `json:"number"`
		Expiry         string              `json:"expiry"`
		SecurityCode   string              `json:"security_code"`
		LastDigits     string              `json:"last_digits"`
		CardType       string              `json:"card_type"`
		BillingAddress *CardBillingAddress `json:"billing_address"`
	}

	// CardBillingAddress structure
	CardBillingAddress struct {
		AddressLine1 string `json:"address_line_1"`
		AddressLine2 string `json:"address_line_2"`
		AdminArea2   string `json:"admin_area_2"`
		AdminArea1   string `json:"admin_area_1"`
		PostalCode   string `json:"postal_code"`
		CountryCode  string `json:"country_code"`
	}

	// PaymentSourceToken structure
	PaymentSourceToken struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	// Payout struct
	Payout struct {
		SenderBatchHeader *SenderBatchHeader `json:"sender_batch_header"`
		Items             []PayoutItem       `json:"items"`
	}

	// PayoutItem struct
	PayoutItem struct {
		RecipientType string        `json:"recipient_type"`
		Receiver      string        `json:"receiver"`
		Amount        *AmountPayout `json:"amount"`
		Note          string        `json:"note,omitempty"`
		SenderItemID  string        `json:"sender_item_id,omitempty"`
	}

	// PayoutItemResponse struct
	PayoutItemResponse struct {
		PayoutItemID      string        `json:"payout_item_id"`
		TransactionID     string        `json:"transaction_id"`
		TransactionStatus string        `json:"transaction_status"`
		PayoutBatchID     string        `json:"payout_batch_id,omitempty"`
		PayoutItemFee     *AmountPayout `json:"payout_item_fee,omitempty"`
		PayoutItem        *PayoutItem   `json:"payout_item"`
		TimeProcessed     *time.Time    `json:"time_processed,omitempty"`
		Links             []Link        `json:"links"`
		Error             ErrorResponse `json:"errors,omitempty"`
	}

	// PayoutResponse struct
	PayoutResponse struct {
		BatchHeader *BatchHeader         `json:"batch_header"`
		Items       []PayoutItemResponse `json:"items"`
		Links       []Link               `json:"links"`
	}

	// RedirectURLs struct
	RedirectURLs struct {
		ReturnURL string `json:"return_url,omitempty"`
		CancelURL string `json:"cancel_url,omitempty"`
	}

	// Refund struct
	Refund struct {
		ID            string                  `json:"id,omitempty"`
		Amount        *Amount                 `json:"amount,omitempty"`
		InvoiceID     string                  `json:"invoice_id,omitempty"`
		CreateTime    PTime                   `json:"create_time,omitempty"`
		Status        RefundStatus            `json:"state,omitempty"`
		NoteToPayer   string                  `json:"note_to_payer"`
		CaptureID     string                  `json:"capture_id,omitempty"`
		ParentPayment string                  `json:"parent_payment,omitempty"`
		UpdateTime    PTime                   `json:"update_time,omitempty"`
		Breakdown     *SellerPayableBreakdown `json:"seller_payable_breakdown"`
	}

	SellerPayableBreakdown struct {
		// gross_amount object
		// The amount that the payee refunded to the payer.
		// Read only.
		GrossAmount *Money `json:"gross_amount,omitempty"`
		// paypal_fee object
		// The PayPal fee that was refunded to the payer. This fee might not match the PayPal fee that the payee paid when the payment was captured.
		// Read only.
		PaypalFee *Money `json:"paypal_fee,omitempty"`
		// net_amount object
		// The net amount that the payee's account is debited, if the payee holds funds in the currency for this refund. The net amount is calculated as gross_amount minus paypal_fee minus platform_fees.
		// Read only.
		NetAmount *Money `json:"net_amount,omitempty"`

		PlatformFees []PlatformFee
	}

	RefundRequest struct {
		Amount      *Amount `json:"amount,omitempty"`
		InvoiceID   string  `json:"invoice_id,omitempty"`
		NoteToPayer string  `json:"note_to_payer,omitempty"`
	}

	// RefundResponse .
	RefundResponse struct {
		ID          string                  `json:"id,omitempty"`
		InvoiceID   string                  `json:"invoice_id,omitempty"`
		Amount      *Amount                 `json:"amount,omitempty"`
		Status      RefundStatus            `json:"status,omitempty"`
		Links       []Link                  `json:"links,omitempty"`
		NoteToPayer string                  `json:"note_to_payer,omitempty"`
		Breakdown   *SellerPayableBreakdown `json:"seller_payable_breakdown,omitempty"`
		CreateTime  PTime                   `json:"create_time,omitempty"`
		UpdateTime  PTime                   `json:"update_time,omitempty"`
	}

	// Related struct
	Related struct {
		Sale          *Sale          `json:"sale,omitempty"`
		Authorization *Authorization `json:"authorization,omitempty"`
		Order         *Order         `json:"order,omitempty"`
		Capture       *Capture       `json:"capture,omitempty"`
		Refund        *Refund        `json:"refund,omitempty"`
	}

	// Sale struct
	Sale struct {
		ID                        string     `json:"id,omitempty"`
		Amount                    *Amount    `json:"amount,omitempty"`
		TransactionFee            *Currency  `json:"transaction_fee,omitempty"`
		Description               string     `json:"description,omitempty"`
		CreateTime                *time.Time `json:"create_time,omitempty"`
		State                     string     `json:"state,omitempty"`
		ParentPayment             string     `json:"parent_payment,omitempty"`
		UpdateTime                *time.Time `json:"update_time,omitempty"`
		PaymentMode               string     `json:"payment_mode,omitempty"`
		PendingReason             string     `json:"pending_reason,omitempty"`
		ReasonCode                string     `json:"reason_code,omitempty"`
		ClearingTime              string     `json:"clearing_time,omitempty"`
		ProtectionEligibility     string     `json:"protection_eligibility,omitempty"`
		ProtectionEligibilityType string     `json:"protection_eligibility_type,omitempty"`
		Links                     []Link     `json:"links,omitempty"`
	}

	// SenderBatchHeader struct
	SenderBatchHeader struct {
		EmailSubject  string `json:"email_subject"`
		SenderBatchID string `json:"sender_batch_id,omitempty"`
	}

	// ShippingAddress struct
	ShippingAddress struct {
		RecipientName string `json:"recipient_name,omitempty"`
		Type          string `json:"type,omitempty"`
		Line1         string `json:"line1"`
		Line2         string `json:"line2,omitempty"`
		City          string `json:"city"`
		CountryCode   string `json:"country_code"`
		PostalCode    string `json:"postal_code,omitempty"`
		State         string `json:"state,omitempty"`
		Phone         string `json:"phone,omitempty"`
	}

	// ShippingDetailAddressPortable used with create orders
	ShippingDetailAddressPortable struct {
		AddressLine1 string `json:"address_line_1,omitempty"`
		AddressLine2 string `json:"address_line_2,omitempty"`
		AdminArea1   string `json:"admin_area_1,omitempty"`
		AdminArea2   string `json:"admin_area_2,omitempty"`
		PostalCode   string `json:"postal_code,omitempty"`
		CountryCode  string `json:"country_code,omitempty"`
	}

	// Name struct
	Name struct {
		FullName string `json:"full_name,omitempty"`
	}

	// ShippingDetail struct
	ShippingDetail struct {
		Name    *Name                          `json:"name,omitempty"`
		Address *ShippingDetailAddressPortable `json:"address,omitempty"`
	}

	expirationTime int64

	// TokenResponse is for API response for the /oauth2/token endpoint
	TokenResponse struct {
		RefreshToken string         `json:"refresh_token"`
		Token        string         `json:"access_token"`
		Type         string         `json:"token_type"`
		ExpiresIn    expirationTime `json:"expires_in"`
	}

	// Transaction struct
	Transaction struct {
		Amount           *Amount         `json:"amount"`
		Description      string          `json:"description,omitempty"`
		ItemList         *ItemList       `json:"item_list,omitempty"`
		InvoiceNumber    string          `json:"invoice_number,omitempty"`
		Custom           string          `json:"custom,omitempty"`
		SoftDescriptor   string          `json:"soft_descriptor,omitempty"`
		RelatedResources []Related       `json:"related_resources,omitempty"`
		PaymentOptions   *PaymentOptions `json:"payment_options,omitempty"`
		NotifyURL        string          `json:"notify_url,omitempty"`
		OrderURL         string          `json:"order_url,omitempty"`
		Payee            *Payee          `json:"payee,omitempty"`
	}

	//Payee struct
	Payee struct {
		Email string `json:"email"`
	}

	// PayeeForOrders struct
	PayeeForOrders struct {
		EmailAddress string `json:"email_address,omitempty"`
		MerchantID   string `json:"merchant_id,omitempty"`
	}

	// UserInfo struct
	UserInfo struct {
		ID              string   `json:"user_id"`
		Name            string   `json:"name"`
		GivenName       string   `json:"given_name"`
		FamilyName      string   `json:"family_name"`
		Email           string   `json:"email"`
		Verified        bool     `json:"verified,omitempty,string"`
		Gender          string   `json:"gender,omitempty"`
		BirthDate       string   `json:"birthdate,omitempty"`
		ZoneInfo        string   `json:"zoneinfo,omitempty"`
		Locale          string   `json:"locale,omitempty"`
		Phone           string   `json:"phone_number,omitempty"`
		Address         *Address `json:"address,omitempty"`
		VerifiedAccount bool     `json:"verified_account,omitempty,string"`
		AccountType     string   `json:"account_type,omitempty"`
		AgeRange        string   `json:"age_range,omitempty"`
		PayerID         string   `json:"payer_id,omitempty"`
	}

	// WebProfile represents the configuration of the payment web payment experience
	//
	// https://developer.paypal.com/docs/api/payment-experience/
	WebProfile struct {
		ID           string       `json:"id,omitempty"`
		Name         string       `json:"name"`
		Presentation Presentation `json:"presentation,omitempty"`
		InputFields  InputFields  `json:"input_fields,omitempty"`
		FlowConfig   FlowConfig   `json:"flow_config,omitempty"`
	}

	// Presentation represents the branding and locale that a customer sees on
	// redirect payments
	//
	// https://developer.paypal.com/docs/api/payment-experience/#definition-presentation
	Presentation struct {
		BrandName  string `json:"brand_name,omitempty"`
		LogoImage  string `json:"logo_image,omitempty"`
		LocaleCode string `json:"locale_code,omitempty"`
	}

	// InputFields represents the fields that are displayed to a customer on
	// redirect payments
	//
	// https://developer.paypal.com/docs/api/payment-experience/#definition-input_fields
	InputFields struct {
		AllowNote       bool `json:"allow_note,omitempty"`
		NoShipping      uint `json:"no_shipping,omitempty"`
		AddressOverride uint `json:"address_override,omitempty"`
	}

	// FlowConfig represents the general behaviour of redirect payment pages
	//
	// https://developer.paypal.com/docs/api/payment-experience/#definition-flow_config
	FlowConfig struct {
		LandingPageType   string `json:"landing_page_type,omitempty"`
		BankTXNPendingURL string `json:"bank_txn_pending_url,omitempty"`
		UserAction        string `json:"user_action,omitempty"`
	}

	VerifyWebhookResponse struct {
		VerificationStatus string `json:"verification_status,omitempty"`
	}

	WebhookEvent struct {
		ID              string          `json:"id"`
		CreateTime      time.Time       `json:"create_time"`
		ResourceType    WHResourceType  `json:"resource_type"`
		EventType       string          `json:"event_type"`
		Summary         string          `json:"summary,omitempty"`
		Resource        json.RawMessage `json:"resource,omitempty"`
		Links           []Link          `json:"links"`
		EventVersion    string          `json:"event_version,omitempty"`
		ResourceVersion string          `json:"resource_version,omitempty"`
	}

	Resource struct {
		// Payment Resource type
		ID                     string                  `json:"id,omitempty"`
		Status                 string                  `json:"status,omitempty"`
		StatusDetails          *CaptureStatusDetails   `json:"status_details,omitempty"`
		Amount                 *PurchaseUnitAmount     `json:"amount,omitempty"`
		UpdateTime             string                  `json:"update_time,omitempty"`
		CreateTime             string                  `json:"create_time,omitempty"`
		ExpirationTime         string                  `json:"expiration_time,omitempty"`
		SellerProtection       *SellerProtection       `json:"seller_protection,omitempty"`
		FinalCapture           bool                    `json:"final_capture,omitempty"`
		SellerPayableBreakdown *CaptureSellerBreakdown `json:"seller_payable_breakdown,omitempty"`
		NoteToPayer            string                  `json:"note_to_payer,omitempty"`
		// merchant-onboarding Resource type
		PartnerClientID string `json:"partner_client_id,omitempty"`
		MerchantID      string `json:"merchant_id,omitempty"`
		// Common
		Links []Link `json:"links,omitempty"`
	}

	CaptureSellerBreakdown struct {
		GrossAmount         PurchaseUnitAmount  `json:"gross_amount"`
		PayPalFee           PurchaseUnitAmount  `json:"paypal_fee"`
		NetAmount           PurchaseUnitAmount  `json:"net_amount"`
		TotalRefundedAmount *PurchaseUnitAmount `json:"total_refunded_amount,omitempty"`
	}

	ReferralRequest struct {
		TrackingID    string      `json:"tracking_id"`
		Operations    []Operation `json:"operations,omitempty"`
		Products      []string    `json:"products,omitempty"`
		LegalConsents []Consent   `json:"legal_consents,omitempty"`
	}

	Operation struct {
		Operation                string              `json:"operation"`
		APIIntegrationPreference *IntegrationDetails `json:"api_integration_preference,omitempty"`
	}

	IntegrationDetails struct {
		RestAPIIntegration *RestAPIIntegration `json:"rest_api_integration,omitempty"`
	}

	RestAPIIntegration struct {
		IntegrationMethod string            `json:"integration_method"`
		IntegrationType   string            `json:"integration_type"`
		ThirdPartyDetails ThirdPartyDetails `json:"third_party_details"`
	}

	ThirdPartyDetails struct {
		Features []string `json:"features"`
	}

	Consent struct {
		Type    string `json:"type"`
		Granted bool   `json:"granted"`
	}

	Tracker struct {
		TransactionID    string         `json:"transaction_id,omitempty"`
		TrackingNumber   string         `json:"tracking_number"`
		Status           TrackingStatus `json:"status,omitempty"`
		Carrier          string         `json:"carrier,omitempty"` // Other
		CarrierNameOther string         `json:"carrier_name_other,omitempty"`
	}

	TrackersRequest struct {
		Trackers []Tracker `json:"trackers"`
	}
	TrackersResponse struct {
		Identifiers []TrackingIdentifier `json:"tracker_identifiers"`
	}

	TrackingIdentifier struct {
		TransactionID  string  `json:"transaction_id,omitempty"`
		TrackingNumber string  `json:"tracking_number,omitempty"`
		Links          []Link  `json:"links"`
		Errors         []Error `json:"errors"`
	}
	Error struct {
		Name            string `json:"name"`
		Message         string `json:"message"`
		DebugID         string `json:"debug_id"`
		InformationLink string `json:"information_link"`
	}
	/*
			{
		  "dispute_id": "PP-D-4012",
		  "create_time": "2019-04-11T04:18:00.000Z",
		  "update_time": "2019-04-21T04:19:08.000Z",
		  "disputed_transactions": [
		    {
		      "seller_transaction_id": "3BC38643YC807283D",
		      "create_time": "2019-04-11T04:16:58.000Z",
		      "transaction_status": "REVERSED",
		      "gross_amount": {
		        "currency_code": "USD",
		        "value": "192.00"
		      },
		      "buyer": {
		        "name": "Lupe Justin"
		      },
		      "seller": {
		        "email": "merchant@example.com",
		        "merchant_id": "5U29WL78XSAEL",
		        "name": "Lesley Paul"
		      }
		    }
		  ],
		  "reason": "MERCHANDISE_OR_SERVICE_NOT_AS_DESCRIBED",
		  "status": "RESOLVED",
		  "dispute_amount": {
		    "currency_code": "USD",
		    "value": "96.00"
		  },
		  "dispute_outcome": {
		    "outcome_code": "RESOLVED_BUYER_FAVOUR",
		    "amount_refunded": {
		      "currency_code": "USD",
		      "value": "96.00"
		    }
		  },
		  "dispute_life_cycle_stage": "CHARGEBACK",
		  "dispute_channel": "INTERNAL",
		  "messages": [
		    {
		      "posted_by": "BUYER",
		      "time_posted": "2019-04-11T04:18:04.000Z",
		      "content": "SNAD case created through automation"
		    }
		  ],
		  "extensions": {
		    "merchandize_dispute_properties": {
		      "issue_type": "SERVICE",
		      "service_details": {
		        "sub_reasons": [
		          "INCOMPLETE"
		        ],
		        "purchase_url": "https://ebay.in"
		      }
		    }
		  },
		  "offer": {
		    "buyer_requested_amount": {
		      "currency_code": "USD",
		      "value": "96.00"
		    }
		  },
		  "links": [
		    {
		      "href": "https://api.sandbox.paypal.com/v1/customer/disputes/PP-D-4012",
		      "rel": "self",
		      "method": "GET"
		    }
		  ]
		}
	*/
	Dispute struct {
		DisputeID      string               `json:"dispute_id,omitempty"`
		CreateTime     PTime                `json:"create_time,omitempty"`
		UpdateTime     PTime                `json:"update_time,omitempty"`
		Transactions   []DisputeTransaction `json:"disputed_transactions,omitempty"`
		Reason         DisputeReason        `json:"reason"`
		Status         DisputeStatus        `json:"status"`
		DisputeAmount  *Money               `json:"dispute_amount"`
		DisputeOutcome struct {
			OutcomeCode    string `json:"outcome_code"`
			AmountRefunded *Money `json:"amount_refunded,omitempty"`
		} `json:"dispute_outcome"`

		Stage          DisputeStage `json:"dispute_life_cycle_stage"`
		DisputeChannel string       `json:"dispute_channel"`
		Messages       []struct {
			PostedBy   string `json:"posted_by"`
			TimePosted PTime  `json:"time_posted"`
			Content    string `json:"content"`
		} `json:"messages"`
		Extensions struct {
			Properties struct {
				IssueType      string `json:"issue_type"`
				ServiceDetails struct {
					SubReasons  []string `json:"sub_reasons"`
					PurchaseURL string   `json:"purchase_url"`
				} `json:"service_details"`
			} `json:"merchandize_dispute_properties"`
		} `json:"extensions"`
		Offer DisputeOffer `json:"offer,omitempty"`
		Links []Link       `json:"links,omitempty"`
	}
	DisputeOffer struct {
		RequestedAmount *Money `json:"buyer_requested_amount"`
	}
	DisputeTransaction struct {
		SellerTransactionID string            `json:"seller_transaction_id,omitempty"`
		CreateTime          PTime             `json:"create_time,omitempty"`
		Status              TransactionStatus `json:"transaction_status,omitempty"`
		InvoiceNumber       string            `json:"invoice_number"`
		GrossAmount         *Money            `json:"gross_amount"`
		Buyer               struct {
			Name string `json:"name"`
		} `json:"buyer"`
		Seller struct {
			Email      string `json:"email,omitempty"`
			MerchantID string `json:"merchant_id,omitempty"`
			Name       string `json:"name,omitempty"`
		} `json:"seller"`
	}
)

// Error method implementation for ErrorResponse struct
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %s", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// MarshalJSON for JSONTime
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`"%s"`, time.Time(t).UTC().Format(time.RFC3339))
	return []byte(stamp), nil
}

func (e *expirationTime) UnmarshalJSON(b []byte) error {
	var n json.Number
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}
	i, err := n.Int64()
	if err != nil {
		return err
	}
	*e = expirationTime(i)
	return nil
}

type PaymentIntent string

const (
	// IntentCapture is CAPTURE. The merchant intends to capture payment immediately
	// after the customer makes a payment.
	IntentCapture PaymentIntent = "CAPTURE"

	// IntentAuthorize is AUTHORIZE. The merchant intends to authorize a payment and
	// place funds on hold after the customer makes a payment.
	// Authorized payments are guaranteed for up to three days
	// but are available to capture for up to 29 days. After the
	// three-day honor period, the original authorized payment
	// expires and you must re-authorize the payment. You must
	// make a separate request to capture payments on demand.
	IntentAuthorize PaymentIntent = "AUTHORIZE"
)

type AuthorizationStatus string

const (
	// AuthorizationStatusCreated is CREATED. The authorized payment is created. No
	// captured payments have been made for this authorized payment.
	AuthorizationStatusCreated AuthorizationStatus = "CREATED"
	// AuthorizationStatusCaptured is CAPTURED. The authorized payment has one or more captures
	// against it. The sum of these captured payments is greater
	// than the amount of the original authorized payment.
	AuthorizationStatusCaptured AuthorizationStatus = "CAPTURED"
	//AuthorizationStatusDenied is DENIED. PayPal cannot authorize funds for this authorized payment.
	AuthorizationStatusDenied AuthorizationStatus = "DENIED"
	//AuthorizationStatusExpired is EXPIRED. The authorized payment has expired.
	AuthorizationStatusExpired AuthorizationStatus = "EXPIRED"
	// AuthorizationStatusPartiallyCaptured is PARTIALLY_CAPTURED. A
	// captured payment was made for the authorized payment for an
	// amount that is less than the amount of the original authorized payment.
	AuthorizationStatusPartiallyCaptured AuthorizationStatus = "PARTIALLY_CAPTURED"
	// AuthorizationStatusVoided is VOIDED. The authorized payment was voided. No more captured
	// payments can be made against this authorized payment.
	AuthorizationStatusVoided AuthorizationStatus = "VOIDED"
	// AuthorizationStatusPending is PENDING. The created authorization is in pending state.
	// For more information, see status.details
	AuthorizationStatusPending AuthorizationStatus = "PENDING"
)

type CaptureStatus string

const (

	// CaptureStatusCompleted is COMPLETED. The funds for this captured payment
	// were credited to the payee's PayPal account.
	CaptureStatusCompleted CaptureStatus = "COMPLETED"

	// CaptureStatusDeclined is DECLINED. The funds could not be captured.
	CaptureStatusDeclined CaptureStatus = "DECLINED"

	// CaptureStatusPartiallyRefunded is PARTIALLY_REFUNDED. An amount less than this captured payment's
	// amount was partially refunded to the payer.
	CaptureStatusPartiallyRefunded CaptureStatus = "PARTIALLY_REFUNDED"

	// CaptureStatusPending isPENDING. The funds for this captured payment was not
	// yet credited to the payee's PayPal account. For more
	// information, see status.details
	CaptureStatusPending CaptureStatus = "PENDING"

	// CaptureStatusRefunded is REFUNDED. An amount greater than or equal to this
	// captured payment's amount was refunded to the payer.
	CaptureStatusRefunded CaptureStatus = "REFUNDED"
)

type OrderStatus string

const (
	// OrderStatusCreated is CREATED. The order was created with the specified context.
	OrderStatusCreated OrderStatus = "CREATED"

	// OrderStatusSaved is SAVED. The order was saved and persisted. The order status
	// continues to be in progress until a capture is made with
	// final_capture = true for all purchase units within the order.
	OrderStatusSaved OrderStatus = "SAVED"

	// OrderStatusApproved is APPROVED. The customer approved the payment through the
	// PayPal wallet or another form of guest or unbranded payment.
	// For example, a card, bank account, or so on.
	OrderStatusApproved OrderStatus = "APPROVED"

	// OrderStatusVoided is VOIDED. All purchase units in the order are voided.
	OrderStatusVoided OrderStatus = "VOIDED"

	// OrderStatusCompleted is COMPLETED. The payment was authorized or the authorized
	// payment was captured for the order.
	OrderStatusCompleted OrderStatus = "COMPLETED"
)

type RefundStatus string

const (
	// RefundStatusCancelled is CANCELLED. The refund was cancelled.
	RefundStatusCancelled RefundStatus = "CANCELLED"
	// RefundStatusPending is PENDING. The refund is pending. For more information, see status_details.reason.
	RefundStatusPending RefundStatus = "PENDING"
	// RefundStatusCompleted is COMPLETED. The funds for this transaction were debited to the customer's account.
	RefundStatusCompleted RefundStatus = "COMPLETED"
)

type DisputeStatus string

const (
	//The dispute is open.
	DisputeStatusOpen DisputeStatus = "OPEN"
	//The dispute is waiting for a response from the customer.
	DisputeStatusWaitingForBuyerResponse DisputeStatus = "WAITING_FOR_BUYER_RESPONSE"
	//The dispute is waiting for a response from the merchant.
	DisputeStatusWaitingForSellerResponse DisputeStatus = "WAITING_FOR_SELLER_RESPONSE"
	//The dispute is under review with PayPal.
	DisputeStatusUnderReview DisputeStatus = "UNDER_REVIEW"
	//The dispute is resolved.
	DisputeStatusResolved DisputeStatus = "RESOLVED"
	//The default status if the dispute does not have one
	//of the other statuses.
	DisputeStatusOther DisputeStatus = "OTHER"
)

type DisputeReason string

const (

	// The customer did not receive the merchandise or service.
	DisputeReasonMerchandiseOrServiceNotReceived DisputeReason = "MERCHANDISE_OR_SERVICE_NOT_RECEIVED"
	// The customer reports that the merchandise or service is not as described.
	DisputeReasonMerchandiseOrServiceNotAsDescribed DisputeReason = "MERCHANDISE_OR_SERVICE_NOT_AS_DESCRIBED"
	// The customer did not authorize purchase of the merchandise or service.
	DisputeReasonUnauthorised DisputeReason = "UNAUTHORISED"
	// The refund or credit was not processed for the customer.
	DisputeReasonCreditNotProcessed DisputeReason = "CREDIT_NOT_PROCESSED"
	// The transaction was a duplicate.
	DisputeReasonDuplicateTransaction DisputeReason = "DUPLICATE_TRANSACTION"
	// The customer was charged an incorrect amount.
	DisputeReasonIncorrectAmount DisputeReason = "INCORRECT_AMOUNT"
	// The customer paid for the transaction through other means.
	DisputeReasonPaymentByOtherMeans DisputeReason = "PAYMENT_BY_OTHER_MEANS"
	// The customer was being charged for a subscription or a recurring transaction that was canceled.
	DisputeReasonCanceledRecurringBilling DisputeReason = "CANCELED_RECURRING_BILLING"
	// A problem occurred with the remittance.
	DisputeReasonProblemWithRemittance DisputeReason = "PROBLEM_WITH_REMITTANCE"
	// Other.
	DisputeReasonOther DisputeReason = "OTHER"
)

type WHResourceType string

const (
	WHResourceTypeCheckout      WHResourceType = "checkout-order"
	WHResourceTypePayment       WHResourceType = "payment"
	WHResourceTypeCapture       WHResourceType = "capture"
	WHResourceTypeOrder         WHResourceType = "order"
	WHResourceTypeRefund        WHResourceType = "refund"
	WHResourceTypeAuthorization WHResourceType = "authorization"
	WHResourceTypeDispute       WHResourceType = "dispute"
)

type DisputeStage string

const (
	// A customer and merchant interact in an attempt to resolve a
	// dispute without escalation to PayPal. Occurs when the customer:
	// Has not received goods or a service.
	// Reports that the received goods or service are not as described.
	// Needs more details, such as a copy of the transaction or a receipt.
	DisputeStageInquiry DisputeStage = "INQUIRY"

	// A customer or merchant escalates an inquiry to a claim,
	// which authorizes PayPal to investigate the case and make
	// a determination. Occurs only when the dispute channel
	// is INTERNAL. This stage is a PayPal dispute lifecycle stage and not a credit card or debit card chargeback. All notes that the customer sends in this stage are visible to PayPal agents only. The customer must wait for PayPal’s response before the customer can take further action. In this stage, PayPal shares dispute details with the merchant, who can complete one of these actions:
	// Accept the claim.
	// Submit evidence to challenge the claim.
	// Make an offer to the customer to resolve the claim.
	DisputeStageChargeback DisputeStage = "CHARGEBACK"

	// The first appeal stage for merchants. A merchant can appeal
	// a chargeback if PayPal's decision is not in the merchant's
	// favor. If the merchant does not appeal within the appeal
	// period, PayPal considers the case resolved.
	DisputeStagePreArbitration DisputeStage = "PRE_ARBITRATION"
	// The second appeal stage for merchants. A merchant can appeal
	// a dispute for a second time if the first appeal was denied.
	// If the merchant does not appeal within the appeal period,
	// the case returns to a resolved status in pre-arbitration stage.
	DisputeStageArbitration DisputeStage = "ARBITRATION"
)

type TransactionStatus string

const (

	// The transaction processing completed.
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	// The items in the transaction are unclaimed. If they are not claimed within 30 days, the funds are returned to the sender.
	TransactionStatusUnclaimed TransactionStatus = "UNCLAIMED"
	// The transaction was denied.
	TransactionStatusDenied TransactionStatus = "DENIED"
	// The transaction failed.
	TransactionStatusFailed TransactionStatus = "FAILED"
	// The transaction is on hold.
	TransactionStatusHeld TransactionStatus = "HELD"
	// The transaction is waiting to be processed.
	TransactionStatusPending TransactionStatus = "PENDING"
	// The payment for the transaction was partially refunded.
	TransactionStatusPartiallyRefunded TransactionStatus = "PARTIALLY_REFUNDED"
	// The payment for the transaction was successfully refunded.
	TransactionStatusRefunded TransactionStatus = "REFUNDED"
	// The payment for the transaction was reversed due to a chargeback or other reversal type.
	TransactionStatusReversed TransactionStatus = "REVERSED"
	// The transaction is cancelled.
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

const (
	// The server returns a minimal response to optimize
	// communication between the API caller and the server.
	// A minimal response includes the id, status and HATEOAS links.
	HeaderPreferMinimal = "return=minimal"
	// The server returns a complete resource representation,
	// including the current state of the resource.
	HeaderPreferRepresentation = "return=representation"
	HeaderPrefer               = "Prefer"
)

type TrackingStatus string

const (
	//The shipment was cancelled and the tracking number no longer applies.
	TrackingStatusCancelled TrackingStatus = "CANCELLED"
	//The item was already delivered when the tracking number was uploaded.
	TrackingStatusDelivered TrackingStatus = "DELIVERED"
	// Either the buyer physically picked up the item or the
	// seller delivered the item in person without involving
	// any couriers or postal companies.
	TrackingStatusLocalPickup TrackingStatus = "LOCAL_PICKUP"
	//The item is on hold. Its shipment was temporarily stopped
	// due to bad weather, a strike, customs, or another reason.
	TrackingStatusOnHold TrackingStatus = "ON_HOLD"
	//The item was shipped and is on the way.
	TrackingStatusShipped TrackingStatus = "SHIPPED"
	//The shipment was created.
	TrackingStatusShipmentCreated TrackingStatus = "SHIPMENT_CREATED"
	//The shipment was dropped off.
	TrackingStatusDroppedOff TrackingStatus = "DROPPED_OFF"
	//The shipment is in transit on its way to the buyer.
	TrackingStatusInTransit TrackingStatus = "IN_TRANSIT"
	//The shipment was returned.
	TrackingStatusReturned TrackingStatus = "RETURNED"
	//The label was printed for the shipment.
	TrackingStatusLabelPrinted TrackingStatus = "LABEL_PRINTED"
	//An error occurred with the shipment.
	TrackingStatusError TrackingStatus = "ERROR"
	//The shipment is unconfirmed.
	TrackingStatusUnconfirmed TrackingStatus = "UNCONFIRMED"
	//Pick-up failed for the shipment.
	TrackingStatusPickupFailed TrackingStatus = "PICKUP_FAILED"
	//The delivery was delayed for the shipment.
	TrackingStatusDeliveryDelayed TrackingStatus = "DELIVERY_DELAYED"
	//The delivery was scheduled for the shipment.
	TrackingStatusDeliveryScheduled TrackingStatus = "DELIVERY_SCHEDULED"
	//The delivery failed for the shipment.
	TrackingStatusDeliveryFailed TrackingStatus = "DELIVERY_FAILED"
	//The shipment is being returned.
	TrackingStatusInreturn TrackingStatus = "INRETURN"
	//The shipment is in process.
	TrackingStatusInProcess TrackingStatus = "IN_PROCESS"
	//The shipment is new.
	TrackingStatusNew TrackingStatus = "NEW"
	//If the shipment is cancelled for any reason, its state is void.
	TrackingStatusVoid TrackingStatus = "VOID"
	//The shipment was processed.
	TrackingStatusProcessed TrackingStatus = "PROCESSED"
	//The shipment was not shipped.
	TrackingStatusNotShipped TrackingStatus = "NOT_SHIPPED"
)
