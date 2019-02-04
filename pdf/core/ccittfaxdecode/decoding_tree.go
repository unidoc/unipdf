package ccittfaxdecode

type decodingTreeNode struct {
	Val    byte
	RunLen *int
	Code   *Code
	Left   *decodingTreeNode
	Right  *decodingTreeNode
}

func addNode(root *decodingTreeNode, code Code, bitPos int, runLen int) {
	val := bitFromUint16(code.Code, bitPos)
	bitPos++

	if val == 1 {
		if root.Right == nil {
			root.Right = &decodingTreeNode{
				Val: val,
			}
		}

		if bitPos == code.BitsWritten {
			root.Right.RunLen = &runLen
			root.Right.Code = &code
		} else {
			addNode(root.Right, code, bitPos, runLen)
		}
	} else {
		if root.Left == nil {
			root.Left = &decodingTreeNode{
				Val: val,
			}
		}

		if bitPos == code.BitsWritten {
			root.Left.RunLen = &runLen
			root.Left.Code = &code
		} else {
			addNode(root.Left, code, bitPos, runLen)
		}
	}
}

func findRunLen(root *decodingTreeNode, code uint16, bitPos int) (*int, *Code) {
	if root == nil {
		return nil, nil
	}

	if bitPos == 16 {
		return root.RunLen, root.Code
	}

	val := bitFromUint16(code, bitPos)
	bitPos++

	var runLenPtr *int
	var codePtr *Code
	if val == 1 {
		runLenPtr, codePtr = findRunLen(root.Right, code, bitPos)
	} else {
		runLenPtr, codePtr = findRunLen(root.Left, code, bitPos)
	}

	if runLenPtr == nil {
		runLenPtr = root.RunLen
		codePtr = root.Code
	}

	return runLenPtr, codePtr
}