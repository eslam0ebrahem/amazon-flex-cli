package config

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

var (
	ApiHost               = "api.amazon.es"
	Endpoint              = "https://api.amazon.es/auth/register"
	RefreshEndpoint       = "https://api.amazon.es/auth/token"
	RabbitEndpoint        = "https://rabbit-eu.amazon.com/RefreshItineraryExternalV3"
	RecordActionsEndpoint = "https://rabbit-eu.amazon.com/RecordItineraryActions"
	AuthDomain            = ".amazon.es"
	DeviceType            = "A1MPSLFC7L5AFK"
	DeviceSerial          = "0f7d079cbd9b4e6e8512568000871bd3"
	DeviceAndroidId       = "a9436c01fd552459"
	AppName               = "com.amazon.flex.rabbit"
	AppVersion            = "314455008"
	SoftwareVersion       = "130050002"
	DeviceModel           = "M2101K7BG"
	OsVersion             = "Redmi/secret_global/secret:12/SP1A.210812.016/V14.0.11.0.TKLMIXM:user/release-keys"
	Manufacturer          = "Xiaomi"
	Product               = "rosemary_global"
	OsVersionShort        = "36"
	UserAgent             = "Dalvik/2.1.0 (Linux; U; Android 16; M2101K7BG Build/BP4A.251205.006)"
	RabbitUserAgent       = "Dalvik/2.1.0 (Linux; U; Android 16; M2101K7BG Build/BP4A.251205.006) RabbitAndroid/3.139.1.14.0"
	RefreshUserAgent      = "AmazonWebView/MAPClientLib/130050002/Android/16/M2101K7BG"
	Frc                   = "AKive7Bzs51BKb+x5b5KaGOfsKatdaBxP9GpfpRexu6t1nA689yL08S28CBDJH5Q2nefvc41F//5rUcJVLyfYhnUpZIbejtgulmZm0mFDu4k39Kp359IBIOcMNMbROc4qZPpt5j8WV8dynGnN60ZuBNntwEIID0ibbREIaiq7ZS4bo6C30VXHM/23SDToonRDPXwqeZuj3igmrtza9/29vSrwfQPtYI6T6nLegaEjzbLKcHVf/59wy/prp9tKF85Ou5ZpV+QVZi0JRIUhgV3iNH1XuPCsq5y0UUjKis1CQZLQ1ZLioveCjh0ytA6RhCxsM90ATkZbvj3ud3dOm4/d99Cg9EgTCV9hAXr+7G2agWLpUeh6raHf9Q0ui4N2jXq/mktg0u7IfY5rdE2Qm6dVvMFUE0Up0PBqx3AVR/1hZZqTajVu9ezY6Xkjba3McjMFvf9AEEXMeHFzH+jK/L28urNZnAMnGrxhQyuyJK2RhrPoM1FlnVeFco="
	MarketplaceId         = "ARBP9OOSHTCHU"
	PreferredLanguage     = "en-US"
)

var (
	DbDir         = "DB"
	TokenFile     = filepath.Join(DbDir, ".flex_tokens.json")
	ItineraryFile = filepath.Join(DbDir, "flex_itinerary.json")
	HistoryFile   = filepath.Join(DbDir, "flex_history.json")
)

func EnsureDbDir() {
	os.MkdirAll(DbDir, 0755)
}

func LoadTokens() map[string]interface{} {
	data, err := os.ReadFile(TokenFile)
	if err != nil {
		return map[string]interface{}{}
	}
	var tokens map[string]interface{}
	json.Unmarshal(data, &tokens)
	return tokens
}

func SaveTokens(data map[string]interface{}, email string, password string) error {
	payload := LoadTokens()
	payload["saved_at"] = time.Now().UTC().Format(time.RFC3339)
	payload["raw_response"] = data

	if email != "" {
		payload["_email"] = base64.StdEncoding.EncodeToString([]byte(email))
	}
	if password != "" {
		payload["_password"] = base64.StdEncoding.EncodeToString([]byte(password))
	}
	
	EnsureDbDir()
	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return os.WriteFile(TokenFile, bytes, 0644)
}

func GetSavedCredentials() (string, string) {
	tokens := LoadTokens()
	email, _ := tokens["_email"].(string)
	password, _ := tokens["_password"].(string)
	
	eBytes, _ := base64.StdEncoding.DecodeString(email)
	pBytes, _ := base64.StdEncoding.DecodeString(password)
	
	return string(eBytes), string(pBytes)
}

func GetBearerToken() string {
	tokens := LoadTokens()

	// ── Path 1: Login response (deep nested) ──────────────────────────────────
	// Written by: api.Login() → config.SaveTokens(loginResp, email, pass)
	// Structure:  raw_response.response.success.tokens.bearer.access_token
	if raw, ok := tokens["raw_response"].(map[string]interface{}); ok {
		if resp, ok := raw["response"].(map[string]interface{}); ok {
			if success, ok := resp["success"].(map[string]interface{}); ok {
				if t, ok := success["tokens"].(map[string]interface{}); ok {
					if bearer, ok := t["bearer"].(map[string]interface{}); ok {
						if at, ok := bearer["access_token"].(string); ok && at != "" {
							return at
						}
					}
				}
			}
		}

		// ── Path 2: Refresh response (flat) ───────────────────────────────────
		// Written by: api.Refresh() → config.SaveTokens(refreshResp, "", "")
		// Structure:  raw_response.access_token
		if at, ok := raw["access_token"].(string); ok && at != "" {
			return at
		}
	}

	return ""
}

func GetDriverName() string {
	tokens := LoadTokens()
	if raw, ok := tokens["raw_response"].(map[string]interface{}); ok {
		if resp, ok := raw["response"].(map[string]interface{}); ok {
			if success, ok := resp["success"].(map[string]interface{}); ok {
				if ext, ok := success["extensions"].(map[string]interface{}); ok {
					if cust, ok := ext["customer_info"].(map[string]interface{}); ok {
						if name, ok := cust["name"].(string); ok {
							return name
						}
					}
				}
			}
		}
	}
	return ""
}


func GetRefreshToken() string {
	tokens := LoadTokens()

	// ── Path 1: Login response (deep nested) ──────────────────────────────────
	if raw, ok := tokens["raw_response"].(map[string]interface{}); ok {
		if resp, ok := raw["response"].(map[string]interface{}); ok {
			if success, ok := resp["success"].(map[string]interface{}); ok {
				if t, ok := success["tokens"].(map[string]interface{}); ok {
					if bearer, ok := t["bearer"].(map[string]interface{}); ok {
						if rt, ok := bearer["refresh_token"].(string); ok && rt != "" {
							return rt
						}
					}
				}
			}
		}

		// ── Path 2: Refresh response (flat) ───────────────────────────────────
		if rt, ok := raw["refresh_token"].(string); ok && rt != "" {
			return rt
		}
	}

	return ""
}

func GetSessionId() string {
	tokens := LoadTokens()
	if sid, ok := tokens["_session_id"].(string); ok {
		return sid
	}
	sid := uuid.New().String()
	tokens["_session_id"] = sid
	EnsureDbDir()
	bytes, _ := json.MarshalIndent(tokens, "", "  ")
	os.WriteFile(TokenFile, bytes, 0644)
	return sid
}
