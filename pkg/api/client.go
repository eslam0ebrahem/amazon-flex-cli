package api

import (
	"errors"
	"fmt"
	"flexcli/pkg/config"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

var client = resty.New().SetTimeout(30 * time.Second)

func init() {
	client.SetHeader("Accept-Encoding", "gzip")
	client.SetHeader("Accept-Language", "en-US")
	client.SetHeader("Connection", "Keep-Alive")
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", config.UserAgent)
}

func Login(email, password string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"auth_data": map[string]interface{}{
			"use_global_authentication": "true",
			"user_id_password": map[string]interface{}{
				"user_id":  email,
				"password": password,
			},
		},
		"registration_data": map[string]interface{}{
			"domain":           "Device",
			"device_type":      config.DeviceType,
			"device_serial":    config.DeviceSerial,
			"app_name":         config.AppName,
			"app_version":      config.AppVersion,
			"device_model":     config.DeviceModel,
			"os_version":       config.OsVersion,
			"software_version": config.SoftwareVersion,
		},
		"requested_token_type": []string{"bearer", "mac_dms", "store_authentication_cookie", "website_cookies"},
		"cookies": map[string]interface{}{
			"domain":          "amazon.es",
			"website_cookies": []string{},
		},
		"age_info": map[string]interface{}{
			"third_party_age_category": "NO_SIGNAL",
		},
		"user_context_map": map[string]interface{}{
			"frc": config.Frc,
		},
		"device_metadata": map[string]interface{}{
			"device_os_family": "android",
			"device_type":      config.DeviceType,
			"device_serial":    config.DeviceSerial,
			"manufacturer":     config.Manufacturer,
			"model":            config.DeviceModel,
			"os_version":       config.OsVersionShort,
			"product":          config.Product,
		},
		"requested_extensions": []string{"device_info", "customer_info"},
	}

	resp, err := client.R().
		SetHeader("Host", config.ApiHost).
		SetHeader("x-amzn-identity-auth-domain", config.AuthDomain).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post(config.Endpoint)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}
	return *resp.Result().(*map[string]interface{}), nil
}

func Refresh() (map[string]interface{}, error) {
	atnr := config.GetRefreshToken()
	if atnr == "" {
		email, password := config.GetSavedCredentials()
		if email != "" && password != "" {
			return Login(email, password)
		}
		return nil, errors.New("no refresh token saved")
	}

	payload := map[string]interface{}{
		"app_name":             config.AppName,
		"app_version":          config.SoftwareVersion,
		"source_token_type":    "refresh_token",
		"source_token":         atnr,
		"requested_token_type": "access_token",
		"device_metadata": map[string]interface{}{
			"device_os_family": "android",
			"device_type":      config.DeviceType,
			"device_serial":    config.DeviceSerial,
			"manufacturer":     config.Manufacturer,
			"model":            config.DeviceModel,
			"os_version":       config.OsVersionShort,
			"product":          config.Product,
		},
		"map_version": map[string]interface{}{
			"current_version":            "20251216N",
			"package_name":               config.AppName,
			"platform":                   "Android",
			"client_metrics_integrated":  true,
		},
		"age_info": map[string]interface{}{
			"third_party_age_category": "NO_SIGNAL",
		},
	}

	resp, err := client.R().
		SetHeader("Host", config.ApiHost).
		SetHeader("User-Agent", config.RefreshUserAgent).
		SetHeader("x-amzn-identity-auth-domain", config.AuthDomain).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post(config.RefreshEndpoint)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}
	return *resp.Result().(*map[string]interface{}), nil
}

func GetUserProfile() (string, error) {
	token := config.GetBearerToken()
	if token == "" {
		return "", errors.New("not logged in")
	}

	resp, err := client.R().
		SetHeader("Host", config.ApiHost).
		SetHeader("x-amzn-identity-auth-domain", config.AuthDomain).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("User-Agent", config.RefreshUserAgent).
		SetResult(&map[string]interface{}{}).
		Get("https://" + config.ApiHost + "/user/profile?attributes=email")

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", errors.New(resp.String())
	}

	result := *resp.Result().(*map[string]interface{})
	if email, ok := result["email"].(string); ok {
		return email, nil
	}
	return "", errors.New("email not found in response")
}

func GetScheduledAssignments() (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("not logged in")
	}

	clientTime := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	instanceId := uuid.New().String()

	resp, err := client.R().
		SetHeader("Host", "flex-capacity-eu.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", clientTime).
		SetHeader("x-flex-instance-id", instanceId).
		SetHeader("x-amzn-marketplace-id", config.MarketplaceId).
		SetResult(&map[string]interface{}{}).
		Get("https://flex-capacity-eu.amazon.com/scheduledAssignments")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return *resp.Result().(*map[string]interface{}), nil
}

func UpdateWorkPhone(phone string) error {
	token := config.GetBearerToken()
	if token == "" {
		return errors.New("not logged in")
	}

	payload := map[string]interface{}{
		"person": map[string]interface{}{
			"workPhoneNumber": phone,
		},
	}

	clientTime := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	instanceId := uuid.New().String()

	resp, err := client.R().
		SetHeader("Host", "tas-uk-extern.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", clientTime).
		SetHeader("x-flex-instance-id", instanceId).
		SetBody(payload).
		Put("https://tas-uk-extern.amazon.com/person")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return errors.New(resp.String())
	}
	return nil
}

func GetAssociatedTRs() (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("not logged in")
	}

	payload := map[string]interface{}{
		"nameValuePairs": map[string]interface{}{},
	}

	resp, err := client.R().
		SetHeader("Host", "ptras-eu-extern.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post("https://ptras-eu-extern.amazon.com/tr/GetAssociatedTRsExternal")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return *resp.Result().(*map[string]interface{}), nil
}

func GetActiveDevice() (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("not logged in")
	}

	clientTime := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	instanceId := uuid.New().String()

	resp, err := client.R().
		SetHeader("Host", "odcs-eu-extern.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", clientTime).
		SetHeader("x-flex-instance-id", instanceId).
		SetBody(map[string]interface{}{}).
		SetResult(&map[string]interface{}{}).
		Post("https://odcs-eu-extern.amazon.com/external/GetActiveDeviceForUserExternal")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return *resp.Result().(*map[string]interface{}), nil
}

func SetActiveDevice() error {
	token := config.GetBearerToken()
	if token == "" {
		return errors.New("not logged in")
	}

	payload := map[string]interface{}{
		"deviceId": config.DeviceSerial,
	}

	clientTime := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	instanceId := uuid.New().String()

	resp, err := client.R().
		SetHeader("Host", "odcs-eu-extern.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", clientTime).
		SetHeader("x-flex-instance-id", instanceId).
		SetBody(payload).
		Post("https://odcs-eu-extern.amazon.com/external/SetActiveDeviceExternal")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return errors.New(resp.String())
	}
	return nil
}

func GetRealTimeAvailability() (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("not logged in")
	}

	clientTime := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	instanceId := uuid.New().String()

	resp, err := client.R().
		SetHeader("Host", "flex-capacity-eu.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("X-Amzn-RequestId", uuid.New().String()).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", clientTime).
		SetHeader("x-flex-instance-id", instanceId).
		SetHeader("x-amzn-app-source", "Google Play Store").
		SetHeader("x-amzn-marketplace-id", config.MarketplaceId).
		SetResult(&map[string]interface{}{}).
		Get("https://flex-capacity-eu.amazon.com/realTimeAvailability")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return *resp.Result().(*map[string]interface{}), nil
}
