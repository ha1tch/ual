package runtime

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

// BringError represents a failed bring operation
type BringError struct {
	Source      *Stack
	Destination *Stack
	Value       []byte
	Reason      string
}

func (e *BringError) Error() string {
	return "bring failed: " + e.Reason
}

// Bring atomically transfers one element from source to destination stack.
// Optional params: for hash destination, first param is key.
// For type conversions, additional params may specify conversion mode (e.g., base for string->int).
func (dest *Stack) Bring(source *Stack, params ...[]byte) error {
	// Lock both stacks (consistent order: source first)
	source.mu.Lock()
	defer source.mu.Unlock()
	dest.mu.Lock()
	defer dest.mu.Unlock()
	
	srcSize := len(source.elements) - source.head
	if srcSize == 0 {
		return &BringError{source, dest, nil, "source stack empty"}
	}
	
	// Determine which element to take based on source's perspective
	var srcIdx int
	switch source.perspective {
	case LIFO:
		srcIdx = len(source.elements) - 1
	case FIFO:
		srcIdx = source.head
	case Indexed, Hash:
		// Default to first valid element
		srcIdx = source.head
	}
	
	srcElem := source.elements[srcIdx]
	srcData := srcElem.data
	
	// Type conversion if needed
	var destData []byte
	var err error
	
	if source.elementType == dest.elementType {
		// Same type, no conversion
		destData = srcData
	} else {
		// Convert based on source and dest types
		destData, err = convert(srcData, source.elementType, dest.elementType, params)
		if err != nil {
			return &BringError{source, dest, srcData, err.Error()}
		}
	}
	
	// Determine key for destination if hash perspective
	var destKey []byte
	if dest.perspective == Hash {
		if len(params) > 0 {
			destKey = params[0]
		} else {
			// No key provided - error for hash destination
			return &BringError{source, dest, srcData, "hash destination requires key"}
		}
	}
	
	// Now we commit: remove from source, add to dest
	// This is the atomic part - we've validated everything
	
	// Remove from source (O(1) for LIFO and FIFO)
	switch source.perspective {
	case LIFO:
		source.elements = source.elements[:srcIdx]
		source.keys = source.keys[:srcIdx]
	case FIFO:
		source.head++
	case Indexed:
		source.elements = append(source.elements[:srcIdx], source.elements[srcIdx+1:]...)
		source.keys = append(source.keys[:srcIdx], source.keys[srcIdx+1:]...)
	case Hash:
		if source.keys[srcIdx] != nil {
			delete(source.hashIdx, string(source.keys[srcIdx]))
		}
		source.elements[srcIdx] = Element{}
		source.keys[srcIdx] = nil
	}
	
	// Add to dest
	newElem := Element{data: destData}
	dest.elements = append(dest.elements, newElem)
	dest.keys = append(dest.keys, destKey)
	if dest.perspective == Hash && destKey != nil {
		dest.hashIdx[string(destKey)] = len(dest.elements) - 1
	}
	
	return nil
}

// convert transforms data from one type to another
func convert(data []byte, from, to ElementType, params [][]byte) ([]byte, error) {
	switch from {
	case TypeInt64:
		return convertFromInt64(data, to, params)
	case TypeFloat64:
		return convertFromFloat64(data, to, params)
	case TypeString:
		return convertFromString(data, to, params)
	case TypeBytes:
		return convertFromBytes(data, to, params)
	case TypeBool:
		return convertFromBool(data, to, params)
	}
	return nil, errors.New("unknown source type")
}

func convertFromInt64(data []byte, to ElementType, params [][]byte) ([]byte, error) {
	val := bytesToInt(data)
	
	switch to {
	case TypeInt64:
		return data, nil
	case TypeFloat64:
		f := float64(val)
		return float64ToBytes(f), nil
	case TypeString:
		base := 10
		if len(params) > 0 {
			base = int(bytesToInt(params[0]))
		}
		s := strconv.FormatInt(val, base)
		return []byte(s), nil
	case TypeBytes:
		return data, nil
	case TypeBool:
		if val == 0 {
			return []byte{0}, nil
		}
		return []byte{1}, nil
	}
	return nil, errors.New("unknown target type")
}

func convertFromFloat64(data []byte, to ElementType, params [][]byte) ([]byte, error) {
	val := bytesToFloat64(data)
	
	switch to {
	case TypeFloat64:
		return data, nil
	case TypeInt64:
		// Default: truncate. Could use params for floor/ceil/round
		return intToBytes(int64(val)), nil
	case TypeString:
		s := strconv.FormatFloat(val, 'f', -1, 64)
		return []byte(s), nil
	case TypeBytes:
		return data, nil
	case TypeBool:
		if val == 0 {
			return []byte{0}, nil
		}
		return []byte{1}, nil
	}
	return nil, errors.New("unknown target type")
}

func convertFromString(data []byte, to ElementType, params [][]byte) ([]byte, error) {
	s := string(data)
	
	switch to {
	case TypeString:
		return data, nil
	case TypeInt64:
		base := 10
		if len(params) > 0 {
			base = int(bytesToInt(params[0]))
		}
		val, err := strconv.ParseInt(s, base, 64)
		if err != nil {
			return nil, errors.New("cannot parse string as integer: " + err.Error())
		}
		return intToBytes(val), nil
	case TypeFloat64:
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, errors.New("cannot parse string as float: " + err.Error())
		}
		return float64ToBytes(val), nil
	case TypeBytes:
		return data, nil
	case TypeBool:
		if s == "true" || s == "1" {
			return []byte{1}, nil
		}
		if s == "false" || s == "0" || s == "" {
			return []byte{0}, nil
		}
		return nil, errors.New("cannot parse string as bool")
	}
	return nil, errors.New("unknown target type")
}

func convertFromBytes(data []byte, to ElementType, params [][]byte) ([]byte, error) {
	switch to {
	case TypeBytes:
		return data, nil
	case TypeInt64:
		return data, nil // interpret bytes as int
	case TypeFloat64:
		return data, nil // interpret bytes as float
	case TypeString:
		return data, nil // bytes are string
	case TypeBool:
		if len(data) == 0 || (len(data) == 1 && data[0] == 0) {
			return []byte{0}, nil
		}
		return []byte{1}, nil
	}
	return nil, errors.New("unknown target type")
}

func convertFromBool(data []byte, to ElementType, params [][]byte) ([]byte, error) {
	val := len(data) > 0 && data[0] != 0
	
	switch to {
	case TypeBool:
		return data, nil
	case TypeInt64:
		if val {
			return intToBytes(1), nil
		}
		return intToBytes(0), nil
	case TypeFloat64:
		if val {
			return float64ToBytes(1.0), nil
		}
		return float64ToBytes(0.0), nil
	case TypeString:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case TypeBytes:
		return data, nil
	}
	return nil, errors.New("unknown target type")
}

// Helper: float64 to bytes
func float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, bits)
	return b
}

// Helper: bytes to float64
func bytesToFloat64(b []byte) float64 {
	if len(b) < 8 {
		padded := make([]byte, 8)
		copy(padded[8-len(b):], b)
		b = padded
	}
	bits := binary.BigEndian.Uint64(b)
	return math.Float64frombits(bits)
}

