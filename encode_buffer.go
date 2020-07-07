package gojay


type Buffer interface {
	Bytes() []byte
	Len() int
	Grow(n int)
	Write(p []byte) (n int, err error)
}

// BufferString adds a string to be encoded, must be used inside a slice or array encoding (does not encode a key)
func (enc *Encoder) BufferString(v Buffer) {
	enc.grow(v.Len() + 4)
	r := enc.getPreviousRune()
	if r != '[' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(string(v.Bytes()))
	enc.writeByte('"')
}

// BufferStringOmitEmpty adds a string to be encoded or skips it if it is zero value.
// Must be used inside a slice or array encoding (does not encode a key)
func (enc *Encoder) BufferStringOmitEmpty(v Buffer) {
	if v == nil || v.Len() == 0 {
		return
	}
	r := enc.getPreviousRune()
	if r != '[' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(string(v.Bytes()))
	enc.writeByte('"')
}

// BufferStringNullEmpty adds a string to be encoded or skips it if it is zero value.
// Must be used inside a slice or array encoding (does not encode a key)
func (enc *Encoder) BufferStringNullEmpty(v Buffer) {
	r := enc.getPreviousRune()
	if v == nil || v.Len() == 0 {
		if r != '[' {
			enc.writeByte(',')
			enc.writeBytes(nullBytes)
		} else {
			enc.writeBytes(nullBytes)
		}
		return
	}
	if r != '[' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(string(v.Bytes()))
	enc.writeByte('"')
}

// BufferStringKey adds a string to be encoded, must be used inside an object as it will encode a key
func (enc *Encoder) BufferStringKey(key string, v Buffer) {
	if enc.hasKeys {
		if !enc.keyExists(key) {
			return
		}
	}
	if v != nil {
		enc.grow(len(key) + v.Len() + 5)
	} else {
		enc.grow(len(key) + 5)
	}

	r := enc.getPreviousRune()
	if r != '{' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(key)
	enc.writeBytes(objKeyStr)
	if v != nil {
		enc.writeStringEscape(string(v.Bytes()))
	}
	enc.writeByte('"')
}

// BufferStringKeyOmitEmpty adds a string to be encoded or skips it if it is zero value.
// Must be used inside an object as it will encode a key
func (enc *Encoder) BufferStringKeyOmitEmpty(key string, v Buffer) {
	if enc.hasKeys {
		if !enc.keyExists(key) {
			return
		}
	}
	if v == nil || v.Len() == 0 {
		return
	}
	enc.grow(len(key) + v.Len() + 5)
	r := enc.getPreviousRune()
	if r != '{' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(key)
	enc.writeBytes(objKeyStr)
	enc.writeStringEscape(string(v.Bytes()))
	enc.writeByte('"')
}

// BufferStringKeyNullEmpty adds a string to be encoded or skips it if it is zero value.
// Must be used inside an object as it will encode a key
func (enc *Encoder) BufferStringKeyNullEmpty(key string, v Buffer) {
	if enc.hasKeys {
		if !enc.keyExists(key) {
			return
		}
	}
	if v != nil {
		enc.grow(len(key) + v.Len() + 5)
	} else {
		enc.grow(len(key) + 5)
	}
	r := enc.getPreviousRune()
	if r != '{' {
		enc.writeTwoBytes(',', '"')
	} else {
		enc.writeByte('"')
	}
	enc.writeStringEscape(key)
	enc.writeBytes(objKey)
	if v == nil || v.Len() == 0 {
		enc.writeBytes(nullBytes)
		return
	}
	enc.writeByte('"')
	if v != nil {
		enc.writeStringEscape(string(v.Bytes()))
	}
	enc.writeByte('"')
}
