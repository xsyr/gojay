package gojay

func (dec *Decoder) decodeBufferString(v Buffer) error {
    for ; dec.cursor < dec.length || dec.read(); dec.cursor++ {
        switch dec.data[dec.cursor] {
        case ' ', '\n', '\t', '\r', ',':
            // is string
            continue
        case '"':
            dec.cursor++
            start, end, err := dec.getString()
            if err != nil {
                return err
            }
            // we do minus one to remove the last quote
            d := dec.data[start : end-1]
            _, err = v.Write(d)
            dec.cursor = end
            return err
        // is nil
        case 'n':
            dec.cursor++
            err := dec.assertNull()
            if err != nil {
                return err
            }
            return nil
        default:
            dec.err = dec.makeInvalidUnmarshalErr(v)
            err := dec.skipData()
            if err != nil {
                return err
            }
            return nil
        }
    }
    return nil
}


// BufferString decodes the JSON value within an object or an array to a *string.
// If next key is not a JSON string nor null, InvalidUnmarshalError will be returned.
func (dec *Decoder) BufferString(v Buffer) error {
    err := dec.decodeBufferString(v)
    if err != nil {
        return err
    }
    dec.called |= 1
    return nil
}
