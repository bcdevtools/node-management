package setup_check

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"sort"
	"strings"
)

func exitWithErrorMsg(error string) {
	printCheckRecords()
	utils.ExitWithErrorMsg(error)
}

func exitWithErrorMsgf(format string, a ...any) {
	printCheckRecords()
	utils.ExitWithErrorMsgf(format, a...)
}

func printCheckRecords() {
	if len(checkRecords) == 0 {
		return
	}

	utils.PrintlnStdErr("\nReports:")

	sort.Slice(checkRecords, func(i, j int) bool {
		left := checkRecords[i]
		right := checkRecords[j]
		if left.fatal && !right.fatal {
			return true
		}
		if !left.fatal && right.fatal {
			return false
		}
		return left.addedNo < right.addedNo
	})

	for idx, record := range checkRecords {
		var sb strings.Builder
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%2d. ", idx+1))
		if record.fatal {
			sb.WriteString("FATAL: ")
		}
		sb.WriteString(record.message)
		if record.suggest != "" {
			sb.WriteString(fmt.Sprintf("\n > %s", record.suggest))
		}
		utils.PrintlnStdErr(sb.String())
	}
}
