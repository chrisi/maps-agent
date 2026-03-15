package camtac

import "fmt"

const (
	indexBitCount    = 12
	lengthBitCount   = 4
	windowSize       = 1 << indexBitCount
	rawLookAheadSize = 1 << lengthBitCount
	breakEven        = (1 + indexBitCount + lengthBitCount) / 9
	lookAheadSize    = rawLookAheadSize + breakEven
	treeRoot         = windowSize
	endOfStream      = 0
	unused           = 0
)

func modWindow(a int) int {
	return a & (windowSize - 1)
}

type treeNode struct {
	parent       int
	smallerChild int
	largerChild  int
}

type context struct {
	window          [windowSize]byte
	tree            [windowSize + 1]treeNode
	dataBuffer      [17]byte
	flagBitMask     int
	bufferOffset    int
	oldBufferOffset int

	compressedSize  int
	incInputString  int
	incOutputString int
}

func (c *context) initTree(r int) {
	for i := 0; i < windowSize+1; i++ {
		c.tree[i] = treeNode{}
	}
	c.tree[treeRoot].largerChild = r
	c.tree[r].parent = treeRoot
	c.tree[r].largerChild = unused
	c.tree[r].smallerChild = unused
}

func (c *context) contractNode(oldNode, newNode int) {
	c.tree[newNode].parent = c.tree[oldNode].parent
	parent := c.tree[oldNode].parent

	if c.tree[parent].largerChild == oldNode {
		c.tree[parent].largerChild = newNode
	} else {
		c.tree[parent].smallerChild = newNode
	}
	c.tree[oldNode].parent = unused
}

func (c *context) replaceNode(oldNode, newNode int) {
	parent := c.tree[oldNode].parent

	if c.tree[parent].smallerChild == oldNode {
		c.tree[parent].smallerChild = newNode
	} else {
		c.tree[parent].largerChild = newNode
	}

	c.tree[newNode] = c.tree[oldNode]

	if c.tree[newNode].smallerChild != unused {
		c.tree[c.tree[newNode].smallerChild].parent = newNode
	}
	if c.tree[newNode].largerChild != unused {
		c.tree[c.tree[newNode].largerChild].parent = newNode
	}

	c.tree[oldNode].parent = unused
}

func (c *context) findNextNode(node int) int {
	next := c.tree[node].smallerChild
	for c.tree[next].largerChild != unused {
		next = c.tree[next].largerChild
	}
	return next
}

func (c *context) deleteString(p int) {
	if c.tree[p].parent == unused {
		return
	}

	if c.tree[p].largerChild == unused {
		c.contractNode(p, c.tree[p].smallerChild)
	} else if c.tree[p].smallerChild == unused {
		c.contractNode(p, c.tree[p].largerChild)
	} else {
		replacement := c.findNextNode(p)
		c.deleteString(replacement)
		c.replaceNode(p, replacement)
	}
}

func (c *context) addString(newNode int, matchPosition *int) int {
	if newNode == endOfStream {
		return 0
	}

	testNode := c.tree[treeRoot].largerChild
	matchLength := 0

	for {
		delta := 0
		i := 0

		for ; i < lookAheadSize; i++ {
			delta = int(c.window[modWindow(newNode+i)]) - int(c.window[modWindow(testNode+i)])
			if delta != 0 {
				break
			}
		}

		if i >= matchLength {
			matchLength = i
			*matchPosition = testNode

			if matchLength >= lookAheadSize {
				c.replaceNode(testNode, newNode)
				return matchLength
			}
		}

		var child *int
		if delta >= 0 {
			child = &c.tree[testNode].largerChild
		} else {
			child = &c.tree[testNode].smallerChild
		}

		if *child == unused {
			*child = newNode
			c.tree[newNode].parent = testNode
			c.tree[newNode].largerChild = unused
			c.tree[newNode].smallerChild = unused
			return matchLength
		}

		testNode = *child
	}
}

func (c *context) initOutputBuffer() {
	c.dataBuffer[0] = 0
	c.flagBitMask = 1
	c.oldBufferOffset = c.bufferOffset
	c.bufferOffset = 1
}

func (c *context) flushOutputBuffer(output *[]byte) bool {
	if c.bufferOffset == 1 {
		return true
	}
	*output = append(*output, c.dataBuffer[:c.bufferOffset]...)
	c.compressedSize += c.bufferOffset
	c.initOutputBuffer()
	return true
}

func (c *context) outputChar(data int, output *[]byte) bool {
	c.dataBuffer[c.bufferOffset] = byte(data)
	c.bufferOffset++
	c.dataBuffer[0] |= byte(c.flagBitMask)
	c.flagBitMask <<= 1
	c.incOutputString = 0

	if c.flagBitMask == 0x100 {
		c.incOutputString = 1
		return c.flushOutputBuffer(output)
	}
	return true
}

func (c *context) outputPair(position, length int, output *[]byte) bool {
	c.dataBuffer[c.bufferOffset] = byte(length << 4)
	c.dataBuffer[c.bufferOffset] |= byte(position >> 8)
	c.bufferOffset++
	c.dataBuffer[c.bufferOffset] = byte(position & 0xff)
	c.bufferOffset++
	c.flagBitMask <<= 1
	c.incOutputString = 0

	if c.flagBitMask == 0x100 {
		c.incOutputString = 1
		return c.flushOutputBuffer(output)
	}
	return true
}

func (c *context) initInputBuffer(inputByte byte) {
	c.flagBitMask = 1
	c.dataBuffer[0] = inputByte
}

func (c *context) inputBit(input []byte, inputIndex *int) (int, error) {
	c.incInputString = 0
	if c.flagBitMask == 0x100 {
		if *inputIndex >= len(input) {
			return 0, fmt.Errorf("unexpected end of input while reading flag byte")
		}
		c.initInputBuffer(input[*inputIndex])
		c.incInputString = 1
	}
	c.flagBitMask <<= 1
	return int(c.dataBuffer[0]) & (c.flagBitMask >> 1), nil
}

// Compress compresses input using the original C implementation's LZSS variant.
func Compress(input []byte) ([]byte, error) {
	var ctxt context
	output := make([]byte, 0, len(input))

	ctxt.compressedSize = 0
	ctxt.initOutputBuffer()

	currentPosition := 1
	inputIndex := 0

	i := 0
	for ; i < lookAheadSize; i++ {
		if inputIndex >= len(input) {
			break
		}
		ctxt.window[currentPosition+i] = input[inputIndex]
		inputIndex++
	}
	lookAheadBytes := i

	ctxt.initTree(currentPosition)

	matchLength := 0
	matchPosition := 0

	for lookAheadBytes > 0 {
		if matchLength > lookAheadBytes {
			matchLength = lookAheadBytes
		}

		var replaceCount int
		if matchLength <= breakEven {
			replaceCount = 1
			if !ctxt.outputChar(int(ctxt.window[currentPosition]), &output) {
				return nil, fmt.Errorf("failed writing literal")
			}
		} else {
			if !ctxt.outputPair(matchPosition, matchLength-(breakEven+1), &output) {
				return nil, fmt.Errorf("failed writing pair")
			}
			replaceCount = matchLength
		}

		for j := 0; j < replaceCount; j++ {
			ctxt.deleteString(modWindow(currentPosition + lookAheadSize))

			if inputIndex >= len(input) {
				lookAheadBytes--
			} else {
				c := input[inputIndex]
				inputIndex++
				ctxt.window[modWindow(currentPosition+lookAheadSize)] = c
			}

			currentPosition = modWindow(currentPosition + 1)
			if lookAheadBytes != 0 {
				matchLength = ctxt.addString(currentPosition, &matchPosition)
			}
		}
	}

	if ctxt.incOutputString == 0 {
		ctxt.flushOutputBuffer(&output)
	}

	return output, nil
}

// Expand expands compressed data.
// outputSize must be the expected decompressed size.
func Expand(input []byte, outputSize int) ([]byte, error) {
	if outputSize < 0 {
		return nil, fmt.Errorf("invalid output size")
	}
	if len(input) == 0 {
		if outputSize == 0 {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("empty input")
	}

	var ctxt context
	output := make([]byte, outputSize)

	inputIndex := 0
	ctxt.initInputBuffer(input[inputIndex])
	inputIndex++

	currentPosition := 1
	outputIndex := 0
	remaining := outputSize

	for remaining > 0 {
		bit, err := ctxt.inputBit(input, &inputIndex)
		if err != nil {
			return nil, err
		}

		if bit != 0 {
			if ctxt.incInputString == 1 {
				inputIndex++
			}
			if inputIndex >= len(input) {
				return nil, fmt.Errorf("unexpected end of input while reading literal")
			}

			c := input[inputIndex]
			inputIndex++

			output[outputIndex] = c
			outputIndex++
			remaining--

			ctxt.window[currentPosition] = c
			currentPosition = modWindow(currentPosition + 1)
		} else {
			if ctxt.incInputString == 1 {
				inputIndex++
			}
			if inputIndex+1 >= len(input) {
				return nil, fmt.Errorf("unexpected end of input while reading pair")
			}

			matchLength := int(input[inputIndex])
			inputIndex++
			matchPosition := int(input[inputIndex])
			inputIndex++

			matchPosition |= (matchLength & 0x0f) << 8
			matchLength >>= 4
			matchLength += breakEven

			if matchLength < remaining {
				remaining -= matchLength + 1
			} else {
				remaining = 0
				matchLength = remaining - 1
			}

			for i := 0; i <= matchLength; i++ {
				c := ctxt.window[modWindow(matchPosition+i)]
				output[outputIndex] = c
				outputIndex++

				ctxt.window[currentPosition] = c
				currentPosition = modWindow(currentPosition + 1)
			}
		}
	}

	return output, nil
}
