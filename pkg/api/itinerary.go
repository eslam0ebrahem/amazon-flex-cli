package api

import (
	"errors"
	"fmt"
	"time"

	"flexcli/pkg/config"

	"github.com/google/uuid"
)

func FetchItinerary(lat, lon float64) (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("no auth token found")
	}

	nowMs := time.Now().UnixMilli()
	nowSec := float64(time.Now().Unix())

	payload := map[string]interface{}{
		"__type":         "RefreshItineraryExternalV3Request:http://internal.amazon.com/coral/com.amazon.rabbit.itinerarycontract/",
		"deviceMetadata": map[string]interface{}{},
		"featureFlags": map[string]interface{}{
			"abReplenishmentFulfillmentEnabled":                 true,
			"amazonLockerEnabled":                               true,
			"anchorPointDeliveryFulfillmentsEnabled":            true,
			"attendedDeliveryFulfillmentsEnabled":               true,
			"bagLevelFulfillmentEnabled":                        true,
			"buildingConsolidationEnabled":                      true,
			"bulkTransportEnabled":                              true,
			"bulkTransportForecastedVolumeEnabled":              true,
			"checkInOperationEnabled":                           true,
			"commingledFulfillmentStopEnabled":                  true,
			"counterHelixEnabled":                               true,
			"eLockerEnabled":                                    true,
			"enrichAddressWithPhotoAttributes":                  true,
			"enrichTrWithEligibleReasonCodes":                   true,
			"exceptionPolicyTypeEnabled":                        true,
			"expectedExecutorsEnabled":                          true,
			"fetchDivertOperations":                             true,
			"fetchPackageNoteDetailsEnabled":                    true,
			"floorDisplayEnabled":                               true,
			"fulfillmentSDPExperienceEnabled":                   true,
			"guidancePhotoAttributesEnabled":                    true,
			"hubDeliveryFulfillmentsEnabled":                    true,
			"kycWorkflowEnabled":                                true,
			"labelAssociationPickupFulfillmentsEnabled":         true,
			"locationInformationWCEnabled":                      true,
			"merchantAdditionalPickupGuardrailsEnabled":         true,
			"onRoadSupplyRunFulfillmentsEnabled":                true,
			"oneTimePasswordDeconsolidationDeprecationEnabled":  true,
			"openBoxDeliveryFulfillmentsEnabled":                true,
			"pickupScanConstraintEnrichmentEnabled":             true,
			"routeTransparencyWorkflowEnabled":                  true,
			"secureDeliveryE2EEnabled":                          true,
			"secureDeliveryWithoutRemovePackagingEnabled":       true,
			"securePickupEnabled":                               true,
			"stationPickupFulfillmentsEnabled":                  true,
			"stopLinkEnabled":                                   true,
			"stopTimeWindowEnabled":                             true,
			"tamperProofTBAREnabled":                            true,
			"trSDPExperienceEnabled":                            true,
			"udsFulfillmentEnabled":                             true,
			"unattendedDeliveryFulfillmentsEnabled":             true,
			"vendAdditionalGeocodes":                            true,
		},
		"itineraryType":            "ON_DUTY_ITINERARY",
		"refreshItineraryMetadata": map[string]interface{}{"pvdHashVersion": "1.0"},
		"refreshToken":             "placeholder",
		"startTransporterSession":  false,
		"transporterContext": map[string]interface{}{
			"marketplaceId":     config.MarketplaceId,
			"preferredLanguage": config.PreferredLanguage,
			"transporterLocation": map[string]interface{}{
				"accuracy":  8.0,
				"altitude":  0.0,
				"latitude":  lat,
				"longitude": lon,
				"provider":  "fused",
				"time":      nowSec,
			},
		},
	}

	resp, err := client.R().
		SetHeader("Host", "rabbit-eu.amazon.com").
		SetHeader("x-amz-access-token", token).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("X-Flex-Client-Time", fmt.Sprintf("%d", nowMs)).
		SetHeader("X-Flex-Instance-Id", config.DeviceSerial).
		SetHeader("X-Amzn-Requestid", uuid.New().String()).
		SetHeader("Accept", "application/vnd.amazon.itinerary+json").
		SetHeader("Version", "1.0").
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post(config.RabbitEndpoint)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}
	return *resp.Result().(*map[string]interface{}), nil
}
