package itinerary

import (
	"encoding/json"
	"flexcli/pkg/config"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

type PackageDetails struct {
	ScannableId        string
	TrId               string
	ActivityId         string
	ActivityType       string
	Status             string
	Reason             string
	ItemId             string
	ClientOrderId      string
	TrackingId         string
	Dimensions         string
	Weight             string
	Latitude           float64
	Longitude          float64
	AddressName        string
	City               string
	Country            string
	Postal             string
	RecipientName      string
	RecipientPhone     string
	DeliveryType       string
	DeliveryInstruct   string
	Images             []string
}

func LoadItinerary() gjson.Result {
	data, err := os.ReadFile(config.ItineraryFile)
	if err != nil {
		return gjson.Result{}
	}
	return gjson.ParseBytes(data)
}

func SaveItinerary(data map[string]interface{}) error {
	config.EnsureDbDir()
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return os.WriteFile(config.ItineraryFile, bytes, 0644)
}

func ExtractActivities(data gjson.Result) []gjson.Result {
	if data.Get("itineraryUpdate.activities").Exists() {
		return data.Get("itineraryUpdate.activities").Array()
	}
	if data.Get("activities").Exists() {
		return data.Get("activities").Array()
	}
	if data.Get("stops").Exists() {
		return data.Get("stops").Array()
	}
	if data.Get("itinerary.activities").Exists() {
		return data.Get("itinerary.activities").Array()
	}
	if data.Get("itinerary.stops").Exists() {
		return data.Get("itinerary.stops").Array()
	}
	return []gjson.Result{}
}

func FindPackage(data gjson.Result, scannableId string) (activity, op, tr, item gjson.Result, found bool) {
	activities := ExtractActivities(data)
	for _, act := range activities {
		ops := act.Get("operations").Array()
		for _, o := range ops {
			trs := o.Get("transportRequests").Array()
			for _, t := range trs {
				items := t.Get("transportItems").Array()
				
				trScannable := t.Get("scannableId").String()
				trackingId := t.Get("clientMetaData.trackingId").String()
				extObjId := t.Get("clientMetaData.externalObjectId").String()
				slamTracking := t.Get("labels.SLAM.details.trackingId.text").String()
				slamAltExec := t.Get("labels.SLAM.details.alternateExecutionId.text").String()

				trScannables := []string{trScannable, trackingId, extObjId, slamTracking, slamAltExec}

				for _, i := range items {
					itemScannable := i.Get("scannableId").String()
					allScannables := append(trScannables, itemScannable)
					for _, s := range allScannables {
						if s != "" && s == scannableId {
							return act, o, t, i, true
						}
					}
				}
			}
		}
	}
	return gjson.Result{}, gjson.Result{}, gjson.Result{}, gjson.Result{}, false
}

func GetPackageLocation(act, op gjson.Result) (float64, float64, bool) {
	if op.Get("deliveryPointGeocode.latitude").Exists() {
		return op.Get("deliveryPointGeocode.latitude").Float(), op.Get("deliveryPointGeocode.longitude").Float(), true
	}
	if op.Get("address.latitude").Exists() {
		return op.Get("address.latitude").Float(), op.Get("address.longitude").Float(), true
	}
	if act.Get("location.latitude").Exists() {
		return act.Get("location.latitude").Float(), act.Get("location.longitude").Float(), true
	}
	
	geocodes := act.Get("activityAddress.geocodes").Array()
	for _, gc := range geocodes {
		if gc.Get("latitude").Exists() {
			return gc.Get("latitude").Float(), gc.Get("longitude").Float(), true
		}
	}
	
	return 0, 0, false
}

func GetPackageImages(tr, item gjson.Result) []string {
	var urls []string
	
	// 1. ivq
	ivqs := tr.Get("itemVerificationQuestions").Array()
	for _, ivq := range ivqs {
		imgUrls := ivq.Get("imageUrls").Map()
		for _, arr := range imgUrls {
			for _, u := range arr.Array() {
				urls = append(urls, u.String())
			}
		}
		imgData := ivq.Get("imageDataMap").Map()
		for _, arr := range imgData {
			for _, obj := range arr.Array() {
				if obj.Get("imageUrl").Exists() {
					urls = append(urls, obj.Get("imageUrl").String())
				}
			}
		}
	}

	// 2. item
	if item.Get("imageUrls").IsArray() {
		for _, u := range item.Get("imageUrls").Array() {
			urls = append(urls, u.String())
		}
	} else {
		if item.Get("imageURL").String() != "" && item.Get("imageURL").String() != "package-url" {
			urls = append(urls, item.Get("imageURL").String())
		}
		if item.Get("imageUrl").String() != "" && item.Get("imageUrl").String() != "package-url" {
			urls = append(urls, item.Get("imageUrl").String())
		}
	}

	// 3. TR
	if tr.Get("imageUrls").IsArray() {
		for _, u := range tr.Get("imageUrls").Array() {
			urls = append(urls, u.String())
		}
	} else {
		if tr.Get("imageURL").String() != "" && tr.Get("imageURL").String() != "package-url" {
			urls = append(urls, tr.Get("imageURL").String())
		}
		if tr.Get("imageUrl").String() != "" && tr.Get("imageUrl").String() != "package-url" {
			urls = append(urls, tr.Get("imageUrl").String())
		}
	}

	seen := make(map[string]bool)
	var deduped []string
	for _, u := range urls {
		if u != "" && u != "package-url" && !seen[u] {
			seen[u] = true
			deduped = append(deduped, u)
		}
	}
	return deduped
}

func GetDetails(scannableId string) (*PackageDetails, error) {
	data := LoadItinerary()
	if !data.Exists() {
		return nil, os.ErrNotExist
	}

	act, op, tr, item, found := FindPackage(data, scannableId)
	if !found {
		return nil, nil
	}

	lat, lon, _ := GetPackageLocation(act, op)
	images := GetPackageImages(tr, item)
	
	addr := op.Get("address")
	if !addr.Exists() {
		addr = act.Get("activityAddress")
	}

	parts := strings.Split(act.Get("activityType").String(), ".")
	actType := parts[len(parts)-1]
	
	state := item.Get("state").String()
	if state == "" {
		state = tr.Get("transportObjectState").String()
		if state == "" {
			state = tr.Get("state").String()
		}
	}

	reason := item.Get("reason").String()
	if reason == "" || reason == "NONE" {
		reason = tr.Get("transportObjectReason").String()
		if reason == "" || reason == "NONE" {
			reason = tr.Get("reason").String()
		}
	}

	// TrId: Amazon's itinerary JSON stores the transport request ID under "id",
	// not "transportRequestId". Try both so we handle any itinerary schema variant.
	trId := tr.Get("id").String()
	if trId == "" {
		trId = tr.Get("transportRequestId").String()
	}

	// ItemId: primary field is "id"; fall back to "itemId".
	itemId := item.Get("id").String()
	if itemId == "" {
		itemId = item.Get("itemId").String()
	}

	// RecipientName: try multiple known field names across itinerary schema versions.
	recipientName := tr.Get("recipientName").String()
	if recipientName == "" {
		recipientName = addr.Get("name").String()
	}

	// RecipientPhone: try multiple known field names.
	recipientPhone := tr.Get("recipientPhoneNumber").String()
	if recipientPhone == "" {
		recipientPhone = tr.Get("recipientPhone").String()
	}
	if recipientPhone == "" {
		recipientPhone = addr.Get("phone").String()
	}

	return &PackageDetails{
		ScannableId:      scannableId,
		TrId:             trId,
		ActivityId:       act.Get("activityId").String(),
		ActivityType:     actType,
		Status:           state,
		Reason:           reason,
		ItemId:           itemId,
		ClientOrderId:    tr.Get("clientMetaData.clientOrderId").String(),
		TrackingId:       tr.Get("clientMetaData.trackingId").String(),
		Dimensions:       tr.Get("trObjectDimensions").Raw,
		Weight:           item.Get("weight").Raw,
		Latitude:         lat,
		Longitude:        lon,
		AddressName:      strings.TrimSpace(addr.Get("addressLine1").String() + " " + addr.Get("addressLine2").String()),
		City:             addr.Get("city").String(),
		Country:          addr.Get("countryCode").String(),
		Postal:           addr.Get("postalCode").String(),
		RecipientName:    recipientName,
		RecipientPhone:   recipientPhone,
		DeliveryType:     tr.Get("contractType").String(),
		DeliveryInstruct: tr.Get("deliveryInstructions").String(),
		Images:           images,
	}, nil
}
