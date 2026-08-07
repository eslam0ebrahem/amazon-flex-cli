package ui

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

func FormatDimensions(raw string) string {
	if !gjson.Valid(raw) {
		return raw
	}

	h := gjson.Get(raw, "height.value").Float()
	hU := gjson.Get(raw, "height.unit").String()
	l := gjson.Get(raw, "length.value").Float()
	lU := gjson.Get(raw, "length.unit").String()
	w := gjson.Get(raw, "width.value").Float()
	wU := gjson.Get(raw, "width.unit").String()

	var parts []string
	if l > 0 {
		parts = append(parts, fmt.Sprintf("%.2f %s (L)", l, lU))
	}
	if w > 0 {
		parts = append(parts, fmt.Sprintf("%.2f %s (W)", w, wU))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%.2f %s (H)", h, hU))
	}

	if len(parts) == 0 {
		return raw
	}
	return strings.Join(parts, " × ")
}

func FormatWeight(raw string) string {
	if !gjson.Valid(raw) {
		return raw
	}
	v := gjson.Get(raw, "value").Float()
	u := gjson.Get(raw, "unit").String()
	if u == "" {
		return raw
	}
	return fmt.Sprintf("%.2f %s", v, u)
}
