package header

import (
	"bytes"
	"fmt"

	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/colors"
)

// PrintHeader prints the header of the application.
func PrintHeader() {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, colors.LightGreen(` __  __     __   __     ______     __         ______     ______     __  __`))
	fmt.Fprintln(&buf, colors.LightGreen(`/\ \/\ \   /\ "-.\ \   /\  ___\   /\ \       /\  __ \   /\  __ \   /\ \/ /`))
	fmt.Fprintln(&buf, colors.Green(`\ \ \_\ \  \ \ \-.  \  \ \ \____  \ \ \____  \ \ \/\ \  \ \  __ \  \ \  _"-.`))
	fmt.Fprintln(&buf, colors.MediumGreen(` \ \_____\  \ \_\\"\_\  \ \_____\  \ \_____\  \ \_____\  \ \_\ \_\  \ \_\ \_\`))
	fmt.Fprintln(&buf, colors.DarkGreen(`  \/_____/   \/_/ \/_/   \/_____/   \/_____/   \/_____/   \/_/\/_/   \/_/\/_/`))
	fmt.Fprintln(&buf, colors.LightGreen("   "+app.Version))
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, app.ShortDesc)
	fmt.Fprintln(&buf, colors.MediumGreen(app.RepoURL))
	fmt.Fprintln(&buf)
	fmt.Print(buf.String())
}
