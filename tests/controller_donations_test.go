package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/guregu/null.v3"

	"github.com/stretchr/testify/assert"
	"twreporter.org/go-api/models"
)

type (
	linePayResultURL struct {
		FrontendRedirectURL string `json:"frontend_redirect_url"`
		BackendRedirectURL  string `json:"backend_redirect_url"`
	}
	donationRecord struct {
		Amount      uint              `json:"amount"`
		CardInfo    models.CardInfo   `json:"card_info"`
		Cardholder  models.Cardholder `json:"cardholder"`
		Currency    string            `json:"currency"`
		Details     string            `json:"details"`
		Frequency   string            `json:"frequency"`
		ID          uint              `json:"id"`
		Notes       string            `json:"notes"`
		OrderNumber string            `json:"order_number"`
		PayMethod   string            `json:"pay_method"`
		SendReceipt string            `json:"send_receipt"`
		ToFeedback  bool              `json:"to_feedback"`
	}
	responseBody struct {
		Status string         `json:"status"`
		Data   donationRecord `json:"data"`
	}
	responseBodyForList struct {
		Status string `json:"status"`
		Data   struct {
			Records []donationRecord `json:"records"`
			Meta    struct {
				Total  uint `json:"total"`
				Offset uint `json:"offset"`
				Limit  uint `json:"limit"`
			}
		} `json:"data"`
	}
	requestBody struct {
		Amount     uint              `json:"amount"`
		Cardholder models.Cardholder `json:"donor"`
		Currency   string            `json:"currency"`
		Details    string            `json:"details"`
		Frequency  string            `json:"frequency"`
		MerchantID string            `json:"merchant_id"`
		PayMethod  string            `json:"pay_method"`
		Prime      string            `json:"prime"`
		ResultURL  linePayResultURL  `json:"result_url"` // Line pay needed only
		UserID     uint              `json:"user_id"`
		ToFeedback bool              `json:"to_feedback"`
	}
)

const (
	testPrime           = "test_3a2fb2b7e892b914a03c95dd4dd5dc7970c908df67a49527c0a648b2bc9"
	testDetails         = "報導者小額捐款"
	testAmount     uint = 500
	testCurrency        = "TWD"
	testMerchantID      = "GlobalTesting_CTBC"
	testFeedback        = true

	testName        = "報導者測試者"
	testAddress     = "台北市南京東路一段32巷100號10樓"
	testNationalID  = "A12345678"
	testPhoneNumber = "+886912345678"
	testZipCode     = "101"

	monthlyFrequency = "monthly"
	yearlyFrequency  = "yearly"
	oneTimeFrequency = "one_time"

	creditCardPayMethod = "credit_card"
)

var testCardholder = models.Cardholder{
	PhoneNumber: null.StringFrom("+886912345678"),
	Name:        null.StringFrom("王小明"),
	Email:       "developer@twreporter.org",
	ZipCode:     null.StringFrom("104"),
	Address:     null.StringFrom("台北市中山區南京東路X巷X號X樓"),
	NationalID:  null.StringFrom("A123456789"),
}

var defaults = struct {
	Total      uint
	Offset     uint
	Limit      uint
	CreditCard string
}{
	Total:      0,
	Offset:     0,
	Limit:      10,
	CreditCard: "credit_card",
}

func testCardholderWithDefaultValue(t *testing.T, ch models.Cardholder) {
	assert.Equal(t, testCardholder.PhoneNumber.ValueOrZero(), ch.PhoneNumber.ValueOrZero())
	assert.Equal(t, testCardholder.Name.ValueOrZero(), ch.Name.ValueOrZero())
	assert.Equal(t, testCardholder.Email, ch.Email)
	assert.Equal(t, testCardholder.ZipCode.ValueOrZero(), ch.ZipCode.ValueOrZero())
	assert.Equal(t, testCardholder.NationalID.ValueOrZero(), ch.NationalID.ValueOrZero())
	assert.Equal(t, testCardholder.Address.ValueOrZero(), ch.Address.ValueOrZero())
}

func testDonationDataValidation(t *testing.T, path string, userID uint, authorization string, cookie http.Cookie) {
	var resp *httptest.ResponseRecorder
	var reqBody requestBody
	var reqBodyInBytes []byte

	t.Run("StatusCode=StatusBadRequest", func(t *testing.T) {
		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of UserID
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			PayMethod: creditCardPayMethod,
			Prime:     testPrime,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of Prime
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			PayMethod: creditCardPayMethod,
			UserID:    userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of Cardholder
		// ===========================================
		reqBody = requestBody{
			Amount:    testAmount,
			PayMethod: creditCardPayMethod,
			Prime:     testPrime,
			UserID:    userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of Cardholder.Email
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Name:        null.StringFrom("王小明"),
				PhoneNumber: null.StringFrom("+886912345678"),
			},
			PayMethod: creditCardPayMethod,
			Prime:     testPrime,
			UserID:    userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of Amount
		// ===========================================
		reqBody = requestBody{
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			PayMethod: creditCardPayMethod,
			Prime:     testPrime,
			UserID:    userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Lack of PayMethod
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			Prime:  testPrime,
			UserID: userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Malformed Email
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer-twreporter,org",
			},
			PayMethod: creditCardPayMethod,
			Prime:     testPrime,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Amount is less than 1(minimum value)
		// ===========================================
		reqBody = requestBody{
			Amount: 0,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			Prime:  testPrime,
			UserID: userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// ===========================================
		// Failure (Data Validation)
		// - Create a Donation by Credit Card
		// - Malformed Cardholder.PhoneNumber (E.164 format)
		// ===========================================
		reqBody = requestBody{
			Amount: 0,
			Cardholder: models.Cardholder{
				Email:       "developer@twreporter.org",
				PhoneNumber: null.StringFrom("0912345678"),
			},
			Prime:  testPrime,
			UserID: userID,
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func testCreateADonationRecord(t *testing.T, path string, userID uint, frequency string, authorization string, cookie http.Cookie) {
	var resp *httptest.ResponseRecorder
	var reqBody requestBody
	var resBody responseBody
	var reqBodyInBytes []byte
	var resBodyInBytes []byte

	// ===========================================
	// Success
	// - Create a Donation by Credit Card
	// - Provide all the fields except `result_url`
	//   in request body
	// ===========================================
	t.Run("StatusCode=StatusCreated", func(t *testing.T) {
		reqBody = requestBody{
			Amount:     testAmount,
			Cardholder: testCardholder,
			Currency:   testCurrency,
			Details:    testDetails,
			MerchantID: testMerchantID,
			Prime:      testPrime,
			UserID:     userID,
		}

		if frequency == oneTimeFrequency {
			reqBody.PayMethod = creditCardPayMethod
		} else {
			reqBody.Frequency = frequency
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
		json.Unmarshal(resBodyInBytes, &resBody)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "success", resBody.Status)
		assert.Equal(t, testAmount, resBody.Data.Amount)
		assert.Equal(t, frequency, resBody.Data.Frequency)
		assert.Equal(t, testCurrency, resBody.Data.Currency)
		assert.Equal(t, testDetails, resBody.Data.Details)
		assert.NotEmpty(t, resBody.Data.OrderNumber)
		assert.Empty(t, resBody.Data.Notes)
		testCardholderWithDefaultValue(t, resBody.Data.Cardholder)

		// ===========================================
		// Success
		// - Create a Donation by Credit Card
		// - Provide minimun required fields
		// ===========================================
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			Frequency:  frequency,
			MerchantID: testMerchantID,
			Prime:      testPrime,
			UserID:     userID,
		}

		if frequency == oneTimeFrequency {
			reqBody.PayMethod = creditCardPayMethod
		} else {
			reqBody.Frequency = frequency
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
		json.Unmarshal(resBodyInBytes, &resBody)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "success", resBody.Status)
		assert.Equal(t, testAmount, resBody.Data.Amount)
		assert.Equal(t, frequency, resBody.Data.Frequency)
		assert.Equal(t, testCurrency, resBody.Data.Currency)
	})

	// ===========================================
	// Failure (Server Error)
	// - Create a Donation by Credit Card
	// - Invalid Prime
	// ===========================================
	t.Run("StatusCode=StatusInternalServerError", func(t *testing.T) {
		reqBody = requestBody{
			Amount: testAmount,
			Cardholder: models.Cardholder{
				Email: "developer@twreporter.org",
			},
			Frequency: frequency,
			Prime:     "test_prime_which_will_occurs_error",
			UserID:    userID,
		}

		if frequency == oneTimeFrequency {
			reqBody.PayMethod = creditCardPayMethod
		} else {
			reqBody.Frequency = frequency
		}

		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("POST", path, string(reqBodyInBytes), "application/json", authorization, cookie)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	// ===========================================
	// Failures (Client Error)
	// - Create a Donation by Credit Card
	// - Request Body Data Validation Error
	// ===========================================
	testDonationDataValidation(t, path, userID, authorization, cookie)
}

func TestCreateADonation(t *testing.T) {
	var authorization, jwt, idToken string
	var cookie http.Cookie
	var path = "/v1/donations/prime"
	var resp *httptest.ResponseRecorder
	var user models.User

	user = getUser(Globs.Defaults.Account)
	jwt = generateJWT(user)
	authorization = fmt.Sprintf("Bearer %s", jwt)

	idToken = generateIDToken(user)
	cookie = http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	// ===========================================
	// Failure (Client Error)
	// - Create a Donation Without Cookie
	// - 401 Unauthorized
	// ===========================================
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTP("POST", path, "", "application/json", "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// ===========================================
	// Failure (Client Error)
	// - Create a Donation Without Authorization Header
	// - 401 Unauthorized
	// ===========================================
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTPWithCookies("POST", path, "", "application/json", "", cookie)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// ===========================================
	// Failure (Client Error)
	// - Create a Donation on Unauthorized Resource
	// - 403 Forbidden
	// ===========================================
	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		resp = serveHTTPWithCookies("POST", path, `{"user_id":1000}`, "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	// ==========================================
	// Test One Time Donation Creation
	// =========================================
	testCreateADonationRecord(t, path, user.ID, oneTimeFrequency, authorization, cookie)
}

func TestCreateAPeriodicDonation(t *testing.T) {
	var authorization, idToken, jwt string
	var cookie http.Cookie
	var path = "/v1/periodic-donations"
	var resp *httptest.ResponseRecorder
	var user models.User

	user = getUser(Globs.Defaults.Account)
	jwt = generateJWT(user)
	authorization = fmt.Sprintf("Bearer %s", jwt)

	idToken = generateIDToken(user)
	cookie = http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	// ===========================================
	// Failure (Client Error)
	// - Create a Periodic Donation Without Authorization Header
	// - 401 Unauthorized
	// ===========================================
	t.Run("StatusCode=StatusUnauthrorized", func(t *testing.T) {
		resp = serveHTTP("POST", path, "", "application/json", "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// ===========================================
	// Failure (Client Error)
	// - Create a Periodic Donation on Unauthenticated Resource
	// - 403 Forbidden
	// ===========================================
	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		resp = serveHTTPWithCookies("POST", path, `{"user_id":1000}`, "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	// ==========================================
	// Test Periodic Donation Creation
	// =========================================
	testCreateADonationRecord(t, path, user.ID, monthlyFrequency, authorization, cookie)
}

func createDefaultDonationRecord(reqBody requestBody, endpoint string, user models.User) responseBody {
	// create jwt of this user
	jwt := generateJWT(user)
	// prepare jwt authorization string
	authorization := fmt.Sprintf("Bearer %s", jwt)

	idToken := generateIDToken(user)
	cookie := http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	// create a donation by HTTP POST request
	reqBodyInBytes, _ := json.Marshal(reqBody)
	resp := serveHTTPWithCookies("POST", endpoint, string(reqBodyInBytes), "application/json", authorization, cookie)
	respInBytes, _ := ioutil.ReadAll(resp.Result().Body)
	defer resp.Result().Body.Close()

	// parse response into struct
	resBody := responseBody{}
	json.Unmarshal(respInBytes, &resBody)
	return resBody
}

func createDefaultPeriodicDonationRecord(user models.User) responseBody {
	// create a default periodic donation record
	path := "/v1/periodic-donations"

	reqBody := requestBody{
		Amount: testAmount,
		Cardholder: models.Cardholder{
			Address:     null.StringFrom(testAddress),
			Email:       user.Email.ValueOrZero(),
			Name:        null.StringFrom(testName),
			NationalID:  null.StringFrom(testNationalID),
			PhoneNumber: null.StringFrom(testPhoneNumber),
			ZipCode:     null.StringFrom(testZipCode),
		},
		Details:    testDetails,
		Frequency:  monthlyFrequency,
		MerchantID: testMerchantID,
		Prime:      testPrime,
		UserID:     user.ID,
		ToFeedback: testFeedback,
	}

	return createDefaultDonationRecord(reqBody, path, user)
}

func createDefaultPrimeDonationRecord(user models.User) responseBody {
	// create a default prime donation record
	path := "/v1/donations/prime"

	reqBody := requestBody{
		Amount: testAmount,
		Cardholder: models.Cardholder{
			Address:     null.StringFrom(testAddress),
			Email:       user.Email.ValueOrZero(),
			Name:        null.StringFrom(testName),
			NationalID:  null.StringFrom(testNationalID),
			PhoneNumber: null.StringFrom(testPhoneNumber),
			ZipCode:     null.StringFrom(testZipCode),
		},
		Details:    testDetails,
		MerchantID: testMerchantID,
		PayMethod:  creditCardPayMethod,
		Prime:      testPrime,
		UserID:     user.ID,
	}

	return createDefaultDonationRecord(reqBody, path, user)
}

func TestPatchAPeriodicDonation(t *testing.T) {
	const donorEmail string = "periodic-donor@twreporter.org"
	var authorization string
	var cookie http.Cookie
	var defaultRecordRes responseBody
	var idToken string
	var jwt string
	var path string
	var reqBody map[string]interface{}
	var reqBodyInBytes []byte
	var resp *httptest.ResponseRecorder
	var user models.User

	// setup before test
	// create a new user
	user = createUser(donorEmail)

	// get record to patch
	defaultRecordRes = createDefaultPeriodicDonationRecord(user)

	jwt = generateJWT(user)
	authorization = fmt.Sprintf("Bearer %s", jwt)
	path = fmt.Sprintf("/v1/periodic-donations/%d", defaultRecordRes.Data.ID)

	idToken = generateIDToken(user)
	cookie = http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	// without cookie
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTP("PATCH", path, "", "application/json", "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// without Authorization header
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTPWithCookies("PATCH", path, "", "application/json", "", cookie)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		var otherUserID uint = 100
		resp = serveHTTPWithCookies("PATCH", path, fmt.Sprintf(`{"user_id": %d}`, otherUserID), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("StatusCode=StatusBadRequest", func(t *testing.T) {
		reqBody = map[string]interface{}{
			"user_id": user.ID,
			// to_feedback should be boolean
			"to_feedback": "false",
			// national_id should be string
			"national_id": true,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("StatusCode=StatusNotFound", func(t *testing.T) {
		var path string
		var recordIDNotFound uint = 1000

		path = fmt.Sprintf("/v1/periodic-donations/%d", recordIDNotFound)

		reqBody = map[string]interface{}{
			"send_receipt": "no",
			"to_feedback":  !testFeedback,
			"user_id":      user.ID,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("StatusCode=StatusNoContent", func(t *testing.T) {
		var dataAfterPatch models.PeriodicDonation

		reqBody = map[string]interface{}{
			"donor": map[string]string{
				"address": "test-addres",
				"name":    "test-name",
			},
			"send_receipt": "no",
			"to_feedback":  !testFeedback,
			"user_id":      user.ID,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		Globs.GormDB.Where("id = ?", defaultRecordRes.Data.ID).Find(&dataAfterPatch)
		assert.Equal(t, reqBody["to_feedback"], dataAfterPatch.ToFeedback.ValueOrZero())
		assert.Equal(t, reqBody["send_receipt"], dataAfterPatch.SendReceipt)
		assert.Equal(t, reqBody["donor"].(map[string]string)["address"], dataAfterPatch.Cardholder.Address.ValueOrZero())
		assert.Equal(t, reqBody["donor"].(map[string]string)["name"], dataAfterPatch.Cardholder.Name.ValueOrZero())
	})
}

func TestPatchAPrimeDonation(t *testing.T) {
	const donorEmail string = "prim-donor@twreporter.org"
	var authorization string
	var cookie http.Cookie
	var defaultRecordRes responseBody
	var idToken string
	var jwt string
	var path string
	var reqBody map[string]interface{}
	var reqBodyInBytes []byte
	var resp *httptest.ResponseRecorder
	var user models.User

	// setup before test
	// create a new user
	user = createUser(donorEmail)

	// get record to patch
	defaultRecordRes = createDefaultPrimeDonationRecord(user)

	jwt = generateJWT(user)
	authorization = fmt.Sprintf("Bearer %s", jwt)
	path = fmt.Sprintf("/v1/donations/prime/%d", defaultRecordRes.Data.ID)

	idToken = generateIDToken(user)
	cookie = http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	// without cookie
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTP("PATCH", path, "", "application/json", "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// without Authorization header
	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		resp = serveHTTPWithCookies("PATCH", path, "", "application/json", "", cookie)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		var otherUserID uint = 100
		resp = serveHTTPWithCookies("PATCH", path, fmt.Sprintf(`{"user_id": %d}`, otherUserID), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("StatusCode=StatusBadRequest", func(t *testing.T) {
		reqBody = map[string]interface{}{
			// national_id should be string
			"donor": map[string]interface{}{
				"national_id": true,
			},
			"user_id": user.ID,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("StatusCode=StatusNotFound", func(t *testing.T) {
		var path string
		var recordIDNotFound uint = 1000

		path = fmt.Sprintf("/v1/donations/prime/%d", recordIDNotFound)
		reqBody = map[string]interface{}{
			"send_receipt": "no",
			"user_id":      user.ID,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("StatusCode=StatusNoContent", func(t *testing.T) {
		var dataAfterPatch models.PayByPrimeDonation

		reqBody = map[string]interface{}{
			"donor": map[string]string{
				"name":    "test-name",
				"address": "test-addres",
			},
			"send_receipt": "no",
			"user_id":      user.ID,
		}
		reqBodyInBytes, _ = json.Marshal(reqBody)
		resp = serveHTTPWithCookies("PATCH", path, string(reqBodyInBytes), "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		Globs.GormDB.Where("id = ?", defaultRecordRes.Data.ID).Find(&dataAfterPatch)
		assert.Equal(t, reqBody["send_receipt"], dataAfterPatch.SendReceipt)
		assert.Equal(t, reqBody["donor"].(map[string]string)["address"], dataAfterPatch.Cardholder.Address.ValueOrZero())
		assert.Equal(t, reqBody["donor"].(map[string]string)["name"], dataAfterPatch.Cardholder.Name.ValueOrZero())
	})
}

func TestGetAPrimeDonationOfAUser(t *testing.T) {
	// setup before test
	donorEmail := "get-prime-donor@twreporter.org"
	// create a new user
	user := createUser(donorEmail)

	primeRes := createDefaultPrimeDonationRecord(user)

	jwt := generateJWT(user)
	authorization := fmt.Sprintf("Bearer %s", jwt)

	idToken := generateIDToken(user)
	cookie := http.Cookie{
		HttpOnly: true,
		MaxAge:   3600,
		Name:     "id_token",
		Secure:   false,
		Value:    idToken,
	}

	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		path := fmt.Sprintf("/v1/donations/prime/%d?user_id=%d", primeRes.Data.ID, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", "", cookie)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		path := fmt.Sprintf("/v1/donations/prime/%d?user_id=%d", primeRes.Data.ID, getUser(Globs.Defaults.Account).ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("StatusCode=StatusNotFound", func(t *testing.T) {
		recordIDNotFound := 1000
		path := fmt.Sprintf("/v1/donations/prime/%d?user_id=%d", recordIDNotFound, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("StatusCode=StatusOK", func(t *testing.T) {
		path := fmt.Sprintf("/v1/donations/prime/%d?user_id=%d", primeRes.Data.ID, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		respInBytes, _ := ioutil.ReadAll(resp.Result().Body)
		defer resp.Result().Body.Close()

		// parse response into struct
		resBody := responseBody{}
		json.Unmarshal(respInBytes, &resBody)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testAmount, resBody.Data.Amount)
		assert.Equal(t, testDetails, resBody.Data.Details)
		assert.Equal(t, donorEmail, resBody.Data.Cardholder.Email)
		assert.Equal(t, testAddress, resBody.Data.Cardholder.Address.ValueOrZero())
		assert.Equal(t, testName, resBody.Data.Cardholder.Name.ValueOrZero())
		assert.Equal(t, testNationalID, resBody.Data.Cardholder.NationalID.ValueOrZero())
		assert.Equal(t, testPhoneNumber, resBody.Data.Cardholder.PhoneNumber.ValueOrZero())
		assert.Equal(t, testZipCode, resBody.Data.Cardholder.ZipCode.ValueOrZero())
		assert.Equal(t, "4242", resBody.Data.CardInfo.LastFour.ValueOrZero())
		assert.Equal(t, "424242", resBody.Data.CardInfo.BinCode.ValueOrZero())
		assert.Equal(t, int64(0), resBody.Data.CardInfo.Funding.ValueOrZero())
		assert.Equal(t, testCurrency, resBody.Data.Currency)
		assert.Equal(t, "credit_card", resBody.Data.PayMethod)
		assert.Equal(t, "monthly", resBody.Data.SendReceipt)
		assert.Equal(t, false, resBody.Data.ToFeedback)
		assert.Equal(t, oneTimeFrequency, resBody.Data.Frequency)
		assert.Empty(t, resBody.Data.Notes)
		assert.NotEmpty(t, resBody.Data.OrderNumber)
	})
}

func TestGetAPeriodicDonationOfAUser(t *testing.T) {
	// setup before test
	donorEmail := "get-periodic-donor@twreporter.org"
	// create a new user
	user := createUser(donorEmail)

	tokenRes := createDefaultPeriodicDonationRecord(user)

	jwt := generateJWT(user)
	authorization := fmt.Sprintf("Bearer %s", jwt)

	idToken := generateIDToken(user)
	cookie := http.Cookie{
		Name:     "id_token",
		Value:    idToken,
		MaxAge:   3600,
		Secure:   false,
		HttpOnly: true,
	}

	t.Run("StatusCode=StatusUnauthorized", func(t *testing.T) {
		path := fmt.Sprintf("/v1/periodic-donations/%d?user_id=%d", tokenRes.Data.ID, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", "", cookie)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("StatusCode=StatusForbidden", func(t *testing.T) {
		path := fmt.Sprintf("/v1/periodic-donations/%d?user_id=%d", tokenRes.Data.ID, getUser(Globs.Defaults.Account).ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("StatusCode=StatusNotFound", func(t *testing.T) {
		recordIDNotFound := 1000
		path := fmt.Sprintf("/v1/periodic-donations/%d?user_id=%d", recordIDNotFound, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("StatusCode=StatusOK", func(t *testing.T) {
		path := fmt.Sprintf("/v1/periodic-donations/%d?user_id=%d", tokenRes.Data.ID, user.ID)
		resp := serveHTTPWithCookies("GET", path, "", "application/json", authorization, cookie)
		assert.Equal(t, http.StatusOK, resp.Code)
		respInBytes, _ := ioutil.ReadAll(resp.Result().Body)
		defer resp.Result().Body.Close()

		// parse response into struct
		resBody := responseBody{}
		json.Unmarshal(respInBytes, &resBody)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, testAmount, resBody.Data.Amount)
		assert.Equal(t, testDetails, resBody.Data.Details)
		assert.Equal(t, donorEmail, resBody.Data.Cardholder.Email)
		assert.Equal(t, testAddress, resBody.Data.Cardholder.Address.ValueOrZero())
		assert.Equal(t, testName, resBody.Data.Cardholder.Name.ValueOrZero())
		assert.Equal(t, testNationalID, resBody.Data.Cardholder.NationalID.ValueOrZero())
		assert.Equal(t, testPhoneNumber, resBody.Data.Cardholder.PhoneNumber.ValueOrZero())
		assert.Equal(t, testZipCode, resBody.Data.Cardholder.ZipCode.ValueOrZero())
		assert.Equal(t, "4242", resBody.Data.CardInfo.LastFour.ValueOrZero())
		assert.Equal(t, "424242", resBody.Data.CardInfo.BinCode.ValueOrZero())
		assert.Equal(t, int64(0), resBody.Data.CardInfo.Funding.ValueOrZero())
		assert.Equal(t, testCurrency, resBody.Data.Currency)
		assert.Equal(t, "monthly", resBody.Data.SendReceipt)
		assert.Equal(t, true, resBody.Data.ToFeedback)
		assert.Equal(t, monthlyFrequency, resBody.Data.Frequency)
		assert.Empty(t, resBody.Data.Notes)
		assert.NotEmpty(t, resBody.Data.OrderNumber)
	})
}

/* GetDonationsOfAUser is not implemented yet
func TestGetDonations(t *testing.T) {
	var resBody responseBodyForList
	var resp *httptest.ResponseRecorder
	var path string

	// set up default records
	setUpBeforeDonationsTest()

	defaultUser := getUser(Globs.Defaults.Account)
	user := getUser(donorEmail)
	jwt := generateJWT(user)
	authorization := fmt.Sprintf("Bearer %s", jwt)

	// ===========================================
	// Failure (Client Error)
	// - Get Donations of A Unkonwn User Without Authorization Header
	// - 401 Unauthorized
	// ===========================================
	path = fmt.Sprintf("/v1/users/%d/donations", user.ID)
	resp = serveHTTP("GET", path, "", "", "")
	assert.Equal(t, 401, resp.Code)

	// ===========================================
	// Failure (Client Error)
	// - Create a Periodic Donation on Unauthenticated Resource
	// - 403 Forbidden
	// ===========================================
	path = fmt.Sprintf("/v1/users/%d/donations", defaultUser.ID)
	resp = serveHTTP("GET", path, "", "", authorization)
	assert.Equal(t, 403, resp.Code)

	// ===========================================
	// Failure (Client Error)
	// - Get Donations of A Unkonwn User
	// - 404 Not Found Error
	// ===========================================
	path = "/v1/users/1000/donations"
	jwt = generateJWT(models.User{
		ID:    1000,
		Email: null.StringFrom("unknown@twreporter.org"),
	})

	resp = serveHTTP("GET", path, "", "", fmt.Sprintf("Bearer %s", jwt))
	assert.Equal(t, 404, resp.Code)

	// ================================================================
	// Success
	// - Get Donations of A User Without `pay_methods` Param
	// - Missing `pay_methods` Param (which means all pay_methods)
	// ================================================================
	path = fmt.Sprintf("/v1/users/%d/donations?offset=%d&limit=%d", user.ID, defaults.Offset, defaults.Limit)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ := ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, defaults.Total, resBody.Data.Meta.Total)
	assert.Equal(t, defaults.Offset, resBody.Data.Meta.Offset)
	assert.Equal(t, defaults.Limit, resBody.Data.Meta.Limit)
	assert.Equal(t, defaults.Total, len(resBody.Data.Records))
	assert.Equal(t, defaults.CreditCard, resBody.Data.Records[0].PayMethod)
	assert.Equal(t, true, resBody.Data.Records[0].IsPeriodic)
	assert.Equal(t, defaults.CreditCard, resBody.Data.Records[1].PayMethod)
	assert.Equal(t, false, resBody.Data.Records[1].IsPeriodic)
	assert.Equal(t, defaults.CreditCard, resBody.Data.Records[2].PayMethod)
	assert.Equal(t, true, resBody.Data.Records[2].IsPeriodic)

	// ===================================================
	// Success
	// - Get Donations of A User Without `offset` Param
	// - Missing `offset` Param (which means offset=0)
	// ===================================================
	path = fmt.Sprintf("/v1/users/%d/donations?pay_methods=credit_card&limit=%d", user.ID, defaults.Limit)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, defaults.Total, resBody.Data.Meta.Total)
	assert.Equal(t, defaults.Offset, resBody.Data.Meta.Offset)

	// =====================================================
	// Success
	// - Get Donations of A User Without `limit` Param
	// - Missing `limit` Param (which means limit=10)
	// =====================================================
	path = fmt.Sprintf("/v1/users/%d/donations?pay_methods=credit_card&offset=%d", user.ID, defaults.Offset)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, defaults.Total, resBody.Data.Meta.Total)
	assert.Equal(t, defaults.Offset, resBody.Data.Meta.Limit)

	// ===================================================
	// Success
	// - Get Donations of A User Without Any Params
	// - Missing `pay_method`, `offset` and `limit` Param
	// ===================================================
	path = fmt.Sprintf("/v1/users/%d/donations", user.ID)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, defaults.Total, resBody.Data.Meta.Total)
	assert.Equal(t, defaults.Limit, resBody.Data.Meta.Limit)
	assert.Equal(t, defaults.Total, len(resBody.Data.Records))

	// ===============================================================
	// Success
	// - Get Donations of A User by Providing `offset=1` and `limit=1`
	// - ?offset=1&limit=1
	// ===============================================================
	path = fmt.Sprintf("/v1/users/%d/donations?offset=1&limit=1", user.ID)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, defaults.Total, resBody.Data.Meta.Total)
	assert.Equal(t, 1, resBody.Data.Meta.Limit)
	assert.Equal(t, 1, len(resBody.Data.Records))
	assert.Equal(t, defaults.CreditCard, resBody.Data.Records[0].PayMethod)
	assert.Equal(t, false, resBody.Data.Records[0].IsPeriodic)

	// ====================================================
	// Success
	// - Get Donations of A User With `offset>total`
	// - ?offset=3&limit=1 (offset is equal to or more than total)
	// ====================================================
	path = fmt.Sprintf("/v1/users/%d/donations?offset=%d&limit=1", user.ID, defaults.Total)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, 0, len(resBody.Data.Records))

	// =========================================================
	// Success
	// - Get Donations of A User With `limit=0`
	// - ?offset=0&limit=0 (limit is 0)
	// =========================================================
	path = fmt.Sprintf("/v1/users/%d/donations?offset=0&limit=0", user.ID)
	resp = serveHTTP("GET", path, "", "", "")
	resBodyInBytes, _ = ioutil.ReadAll(resp.Result().Body)
	json.Unmarshal(resBodyInBytes, &resBody)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, 0, len(resBody.Data.Records))

	// =========================================================
	// Success
	// - Get Donations of A User
	// - Test offset and limit are not unsigned integer
	// - Test SQL Injection, put statement in pay_methods
	// - ?limit=NaN&offset=-1&pay_methods=;select * from users;
	// =========================================================
	path = fmt.Sprintf("/v1/users/%d/donations?limit=NaN&offset=-1&pay_methods=;select * from users;", user.ID)
	resp = serveHTTP("GET", path, "", "", "")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "success", resBody.Status)
	assert.Equal(t, 3, len(resBody.Data.Records))
}
*/
