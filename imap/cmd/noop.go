package cmd

type Noop struct {
}

// func (this *Noop) cmdNoop(input string) bool {
// 	inputN := strings.SplitN(input, " ", 2)
// 	if len(inputN) == 2 {
// 		if this.cmdCompare(inputN[1], CMD_NOOP) {
// 			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
// 			return true
// 		}
// 	}
// 	return false
// }
