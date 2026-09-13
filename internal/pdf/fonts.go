// internal/pdf/fonts.go
package pdf

import (
	_ "embed"
	"encoding/base64"
)

//go:embed assets/fonts/Vazirmatn-Regular.woff2
var vazirRegular []byte

//go:embed assets/fonts/Vazirmatn-Bold.woff2
var vazirBold []byte

func FontFaceCSS() string {
	return `
	@font-face {
		font-family: "Vazirmatn";
		src: url(data:font/woff2;base64,` + base64.StdEncoding.EncodeToString(vazirRegular) + `) format("woff2");
		font-weight: 400;
		font-display: swap;
	}
	@font-face {
		font-family: "Vazirmatn";
		src: url(data:font/woff2;base64,` + base64.StdEncoding.EncodeToString(vazirBold) + `) format("woff2");
		font-weight: 700;
		font-display: swap;
	}`
}
