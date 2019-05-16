/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

// decodingTreeNode is a node of a tree which represents the finite state machine for
// searching the decoded pixel run lengths having the bit sequence. `Val` is a single bit
// in a byte value. It is used to navigate the tree. Each node either contains a decoded
// run length and the corresponding code or not.
type decodingTreeNode struct {
	Val    byte
	RunLen *int
	Code   *code
	Left   *decodingTreeNode
	Right  *decodingTreeNode
}

// addNode adds the node to the decoding tree.
func addNode(root *decodingTreeNode, code code, bitPos int, runLen int) {
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

// findRunLen searches for the decoded pixel run length and the corresponding code
// having the bit sequence represented as a single uint16 value.
func findRunLen(root *decodingTreeNode, codeNum uint16, bitPos int) (*int, *code) {
	if root == nil {
		return nil, nil
	}

	if bitPos == 16 {
		return root.RunLen, root.Code
	}

	val := bitFromUint16(codeNum, bitPos)
	bitPos++

	var runLenPtr *int
	var codePtr *code
	if val == 1 {
		runLenPtr, codePtr = findRunLen(root.Right, codeNum, bitPos)
	} else {
		runLenPtr, codePtr = findRunLen(root.Left, codeNum, bitPos)
	}

	if runLenPtr == nil {
		runLenPtr = root.RunLen
		codePtr = root.Code
	}

	return runLenPtr, codePtr
}
