package results

import (
	"golang.org/x/text/language"
	"strings"
)

// x/text includes historical and CLDR-only territories. Exclude those even
// when CLDR supplies a three-letter code; private XK is not an ISO assignment.
const nonISOCountries = " AC AN BU CP CS CT DD DG DY FQ FX HV JT MI NH NQ NT PC PU PZ RH SU TA TP VD WK XK YD YU ZR "

func validCountry(code string) bool {
	if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' || strings.Contains(nonISOCountries, " "+code+" ") {
		return false
	}
	region, err := language.ParseRegion(code)
	return err == nil && region.String() == code && region.IsCountry() && region.ISO3() != "ZZZ"
}
