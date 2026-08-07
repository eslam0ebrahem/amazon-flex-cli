package ui

import (
	"regexp"

	"github.com/01walid/goarabic"
)

// FixArabic handles shaping and reversing Arabic text so that it renders 
// correctly left-to-right in terminal emulators without breaking UI boxes/padding.
func FixArabic(s string) string {
	// Match contiguous blocks of Arabic letters and spaces
	re := regexp.MustCompile(`[\p{Arabic}\s]+`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Only shape and reverse if it actually contains Arabic
		hasArabic := false
		for _, r := range match {
			if r >= 0x0600 && r <= 0x06FF {
				hasArabic = true
				break
			}
		}
		if !hasArabic {
			return match
		}
		
		reshaped := goarabic.ToGlyph(match)
		reversed := goarabic.Reverse(reshaped)
		return reversed
	})
}
