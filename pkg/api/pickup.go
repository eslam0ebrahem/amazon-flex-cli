package api

import (
	"errors"
	"flexcli/pkg/config"

	"github.com/google/uuid"
)

func BuildPickupPayload(scannableId, trId, itemId string, lat, lon float64, timestamp float64) map[string]interface{} {
	sessionId := config.GetSessionId()
	actionId1 := uuid.New().String()
	actionId2 := uuid.New().String()

	return map[string]interface{}{
		"__type":             "RecordItineraryActionsExternalRequest:http://internal.amazon.com/coral/com.amazon.fips/",
		"enrichmentUpdates":  []interface{}{},
		"fulfillmentUpdates": []interface{}{},
		"itineraryActions": []interface{}{
			map[string]interface{}{
				"__type":              "PackageUpdateAction:http://internal.amazon.com/coral/com.amazon.fips/",
				"actionId":            actionId1,
				"actionType":          "PACKAGE_AND_ITEM_UPDATE",
				"deliverySignatureId": "",
				"isExpressPickup":     false,
				"packageDetails": map[string]interface{}{
					"deliveryDetails": map[string]interface{}{
						"pickupTimestamp":     timestamp,
						"recipientDoorNumber": "",
						"recipientName":       "",
					},
					"itemUpdateList": []interface{}{},
					"location": map[string]interface{}{
						"accuracy":  8.0,
						"altitude":  0.0,
						"latitude":  lat,
						"longitude": lon,
						"provider":  "fused",
						"time":      timestamp,
					},
					"scannableId":    scannableId,
					"trId":           trId,
					"trObjectReason": "NONE",
					"trObjectState":  "PICKED_UP",
				},
				"packageUpdateActionMetadata":  map[string]interface{}{"SessionId": sessionId},
				"timestamp":                    timestamp,
				"workflowConstructTypeDataMap": map[string]interface{}{},
			},
			map[string]interface{}{
				"__type":          "ItemUpdateAction:http://internal.amazon.com/coral/com.amazon.fips/",
				"actionId":        actionId2,
				"actionType":      "PACKAGE_AND_ITEM_UPDATE",
				"isExpressPickup": false,
				"itemDetails": map[string]interface{}{
					"itemId":          itemId,
					"reason":          "NONE",
					"state":           "PICKED_UP",
					"updateTimestamp": timestamp,
				},
				"scannableId": scannableId,
				"timestamp":   timestamp,
				"trId":        trId,
			},
		},
		"recordActionsMetadata": map[string]interface{}{
			"transporterMetadata": map[string]interface{}{
				"marketplaceId": config.MarketplaceId,
			},
		},
	}
}

func BuildUpdatePayload(scannableId, trId, itemId string, lat, lon float64, timestamp float64, reason, state string) map[string]interface{} {
	sessionId := config.GetSessionId()
	actionId1 := uuid.New().String()
	actionId2 := uuid.New().String()

	return map[string]interface{}{
		"__type":             "RecordItineraryActionsExternalRequest:http://internal.amazon.com/coral/com.amazon.fips/",
		"enrichmentUpdates":  []interface{}{},
		"fulfillmentUpdates": []interface{}{},
		"itineraryActions": []interface{}{
			map[string]interface{}{
				"__type":              "PackageUpdateAction:http://internal.amazon.com/coral/com.amazon.fips/",
				"actionId":            actionId1,
				"actionType":          "PACKAGE_AND_ITEM_UPDATE",
				"deliverySignatureId": "",
				"isExpressPickup":     false,
				"packageDetails": map[string]interface{}{
					"deliveryDetails": map[string]interface{}{
						"pickupTimestamp":     timestamp,
						"recipientDoorNumber": "",
						"recipientName":       "",
					},
					"itemUpdateList": []interface{}{},
					"location": map[string]interface{}{
						"accuracy":  8.0,
						"altitude":  0.0,
						"latitude":  lat,
						"longitude": lon,
						"provider":  "fused",
						"time":      timestamp,
					},
					"scannableId":    scannableId,
					"trId":           trId,
					"trObjectReason": reason,
					"trObjectState":  state,
				},
				"packageUpdateActionMetadata":  map[string]interface{}{"SessionId": sessionId},
				"timestamp":                    timestamp,
				"workflowConstructTypeDataMap": map[string]interface{}{},
			},
			map[string]interface{}{
				"__type":          "ItemUpdateAction:http://internal.amazon.com/coral/com.amazon.fips/",
				"actionId":        actionId2,
				"actionType":      "PACKAGE_AND_ITEM_UPDATE",
				"isExpressPickup": false,
				"itemDetails": map[string]interface{}{
					"itemId":          itemId,
					"reason":          "NONE",
					"state":           state,
					"updateTimestamp": timestamp,
				},
				"scannableId": scannableId,
				"timestamp":   timestamp,
				"trId":        trId,
			},
		},
		"recordActionsMetadata": map[string]interface{}{
			"transporterMetadata": map[string]interface{}{
				"marketplaceId": config.MarketplaceId,
			},
		},
	}
}

func RecordPickup(payload map[string]interface{}) (map[string]interface{}, error) {
	token := config.GetBearerToken()
	if token == "" {
		return nil, errors.New("no auth token found")
	}

	resp, err := client.R().
		SetHeader("x-amz-access-token", token).
		SetHeader("User-Agent", config.RabbitUserAgent).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&map[string]interface{}{}).
		Post(config.RecordActionsEndpoint)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}
	return *resp.Result().(*map[string]interface{}), nil
}
