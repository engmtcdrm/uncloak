package header

import (
	"fmt"

	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/colors"
)

// PrintHeader prints the header of the application.
func PrintHeader() {
	fmt.Println(colors.LightGreen(` __  __     __   __     ______     __         ______     ______     __  __`))
	fmt.Println(colors.LightGreen(`/\ \/\ \   /\ "-.\ \   /\  ___\   /\ \       /\  __ \   /\  __ \   /\ \/ /`))
	fmt.Println(colors.Green(`\ \ \_\ \  \ \ \-.  \  \ \ \____  \ \ \____  \ \ \/\ \  \ \  __ \  \ \  _"-.`))
	fmt.Println(colors.MediumGreen(` \ \_____\  \ \_\\"\_\  \ \_____\  \ \_____\  \ \_____\  \ \_\ \_\  \ \_\ \_\`))
	fmt.Println(colors.DarkGreen(`  \/_____/   \/_/ \/_/   \/_____/   \/_____/   \/_____/   \/_/\/_/   \/_/\/_/`))
	fmt.Println(colors.LightGreen("   " + app.Version))
	fmt.Println()
	fmt.Println(app.ShortDesc)
	fmt.Println(colors.MediumGreen(app.RepoUrl))
	fmt.Println()
}
