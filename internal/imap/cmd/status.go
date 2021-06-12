package cmd

type Status struct {
}

// func (this *Status) cmdStatus(input string) bool {
// 	inputN := strings.SplitN(input, " ", 4)
// 	if len(inputN) == 4 {
// 		if this.cmdCompare(inputN[1], CMD_STATUS) {
// 			this.selectBox = strings.Trim(inputN[2], "\"")
// 			outArgs := this.parseArgs(inputN[3])
// 			this.writeArgs("* %s %s %s", inputN[1], inputN[2], outArgs)
// 			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
// 			return true
// 		}
// 	}
// 	return false
// }
