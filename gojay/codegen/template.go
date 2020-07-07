package codegen

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	decodeBaseType = iota
	encodeBaseType
	decodeBaseTypeSlice
	encodeBaseTypeSlice
	decodeRawType
	encodeRawType

	decodeByteSliceAsString
	encodeByteSliceAsString

	decodeStruct
	decodeAsBuffer
	decodeAsBufferString

	encodeStruct
	encodeAsBuffer
	encodeAsBufferString

	decodeStructSlice
	encodeStructSlice
	decodeTime
	encodeTime

	decodeSQLNull
	encodeSQLNull

	decodeUnknown
	encodeUnknown

	resetFieldValue
	poolInstanceRelease
	poolSliceInstanceRelease
)

var fieldTemplate = map[int]string{
	decodeBaseType: `		case "{{.Key}}":
{{if .IsPointer}}			var value {{.Type}}
			err := dec.{{.DecodingMethod}}(&value)
			if err == nil {
				{{.Accessor}} = &value
			}
			return err
{{else}} 		return dec.{{.DecodingMethod}}(&{{.Accessor}}){{end}}
`,
	encodeBaseType: `    enc.{{.EncodingMethod}}Key{{.OmitEmpty}}("{{.Key}}", {{.DereferenceModifier}}{{.Accessor}})`,

	encodeByteSliceAsString: `    enc.StringKey{{.OmitEmpty}}("{{.Key}}", string({{.DereferenceModifier}}{{.Accessor}}))`,

	decodeBaseTypeSlice: `		case "{{.Key}}":
			var aSlice = {{.HelperType}}{}
			err := dec.Array(&aSlice)
			if err == nil && len(aSlice) > 0 {
				{{.Mutator}} = {{.RawType}}(aSlice)
			}
			return err
`,
	encodeBaseTypeSlice: `    var {{.Var}}Slice = {{.HelperType}}({{.Accessor}})
    enc.ArrayKey{{.OmitEmpty}}("{{.Key}}",{{.Var}}Slice)`,

	decodeRawType: `		case "{{.Key}}":
			var value = gojay.EmbeddedJSON{}
			err := dec.AddEmbeddedJSON(&value)
			if err == nil && len(value) > 0 {
				{{.Mutator}} = {{.Type}}(value)
			}
			return err
`,

	encodeRawType: `    var {{.Var}}Slice = gojay.EmbeddedJSON({{.Accessor}})
    enc.AddEmbeddedJSONKey{{.OmitEmpty}}("{{.Key}}", &{{.Var}}Slice)`,
	decodeStruct: `		case "{{.Key}}":{{if .IsPointer}}
			var value = {{.Init}}
			err := dec.Object(value)
			if err == nil {
				{{.Mutator}} = value
			}
{{else}}
			err := dec.Object(&{{.Mutator}})
{{end}}
			return err
`,
	decodeAsBuffer: `		case "{{.Key}}":{{if .IsPointer}}
			var value = {{.Init}}
			err := dec.Object(value)
			if err == nil {
				{{.Mutator}} = value
			}
{{else}}
			err := dec.Object(&{{.Mutator}})
{{end}}
			return err
`,
	decodeAsBufferString: `		case "{{.Key}}":{{if .IsPointer}}
			var value = {{.Init}}
			err := dec.BufferString(value)
			if err == nil {
				{{.Mutator}} = value
			}
{{else}}
			err := dec.Object(&{{.Mutator}})
{{end}}
			return err
`,
	encodeStruct: `    enc.ObjectKey{{.OmitEmpty}}("{{.Key}}", {{.PointerModifier}}{{.Accessor}})`,
	encodeAsBuffer: `    enc.ObjectKey{{.OmitEmpty}}("{{.Key}}", {{.PointerModifier}}{{.Accessor}})`,
	encodeAsBufferString: `    enc.BufferStringKey{{.OmitEmpty}}("{{.Key}}", {{.PointerModifier}}{{.Accessor}})`,

	decodeStructSlice: `		case "{{.Key}}":
			   var aSlice = {{.HelperType}}{}
			   err := dec.Array(&aSlice)
			   if err == nil && len(aSlice) > 0 {
				   {{.Mutator}} = {{.RawType}}(aSlice)
			   }
			   return err
   `,

	encodeStructSlice: `    var {{.Var}}Slice = {{.HelperType}}({{.Accessor}})
    enc.ArrayKey{{.OmitEmpty}}("{{.Key}}", {{.DereferenceModifier}}{{.Var}}Slice)`,

	decodeTime: `		case "{{.Key}}":
			var format = {{.TimeLayout}}
			var value = {{.Init}}
			err := dec.Time({{.PointerModifier}}value, format)
			if err == nil {
				{{.Mutator}} = value
			}
			return err
`,

	encodeTime: `{{if .IsPointer}}    if {{.Accessor}} != nil {
        enc.TimeKey("{{.Key}}", {{.PointerModifier}}{{.Accessor}}, {{.TimeLayout}})
    }{{else}}    enc.TimeKey("{{.Key}}", {{.PointerModifier}}{{.Accessor}}, {{.TimeLayout}}){{end}}`,
	decodeSQLNull: `		case "{{.Key}}":
			var value = {{.Init}}
			err := dec.SQLNull{{.NullType}}({{.PointerModifier}}value)
			if err == nil {
				{{.Mutator}} = value
			}
			return err
`,
	encodeSQLNull: `{{if .IsPointer}}    if {{.Accessor}} != nil {
        enc.SQLNull{{.NullType}}Key{{.OmitEmpty}}("{{.Key}}", {{.PointerModifier}}{{.Accessor}})
    }{{else}}    enc.SQLNull{{.NullType}}Key{{.OmitEmpty}}("{{.Key}}", {{.PointerModifier}}{{.Accessor}}){{end}}`,
	decodeUnknown: `		case "{{.Key}}":
			return dec.Any({{.PointerModifier}}{{.Accessor}})
`,
	encodeUnknown: `{{if .IsPointer}}    if {{.Accessor}} != nil {	
		enc.Any({{.Accessor}})
	}{{else}}enc.Any({{.Accessor}}){{end}}`,
	resetFieldValue: `{{if .ResetDependency}}{{.ResetDependency}}
{{end}}    {{.Mutator}} = {{.Reset}}`,
	poolInstanceRelease: `	{{.PoolName}}.Put({{.Accessor}})`,

	poolSliceInstanceRelease: `	for i := range {{.Accessor}} {
        {{.Accessor}}[i].Reset()
		{{.PoolName}}.Put({{.PointerModifier}}{{.Accessor}}[i])
    }`,
}

const (
	fileCode = iota
	encodingStructType
	baseTypeSlice
	structTypeSlice
	resetStruct
	poolVar
	poolInit
	embeddedStructInit
	timeSlice
	typeSlice
)

var blockTemplate = map[int]string{
	fileCode: `// Code generated by Gojay. DO NOT EDIT.


package {{.Pkg}}

import (
	{{.Imports}}
)
{{if .Init}}
func init() {
{{.Init}}
}
{{end}}
{{.Code}}
`,
	encodingStructType: `// MarshalJSONObject implements MarshalerJSONObject
func ({{.Receiver}}) MarshalJSONObject(enc *gojay.Encoder) {
{{.EncodingCases}}
}

// IsNil checks if instance is nil
func ({{.Receiver}}) IsNil() bool {
    return {{.Alias}} == nil
}

// UnmarshalJSONObject implements gojay's UnmarshalerJSONObject
func ({{.Receiver}}) UnmarshalJSONObject(dec *gojay.Decoder, k string) error {
{{.InitEmbedded}}
	switch k {
{{.DecodingCases}}	
	}
	return nil
}

// NKeys returns the number of keys to unmarshal
func ({{.Receiver}}) NKeys() int { return {{.FieldCount}} }

{{.Reset}}

`,

	baseTypeSlice: `

type {{.HelperType}} {{.RawType}}

// UnmarshalJSONArray decodes JSON array elements into slice
func (a *{{.HelperType}}) UnmarshalJSONArray(dec *gojay.Decoder) error {
	var value {{.ComponentType}}
	if err := dec.{{.DecodingMethod}}(&value); err != nil {
		return err
	}
	*a = append(*a, {{.ComponentInitModifier}}value)
	return nil
}

// MarshalJSONArray encodes arrays into JSON
func (a {{.HelperType}}) MarshalJSONArray(enc *gojay.Encoder) {
	for _, item := range a {
		enc.{{.EncodingMethod}}({{.ComponentDereferenceModifier}}item)
	}
}

// IsNil checks if array is nil
func (a {{.HelperType}}) IsNil() bool {
	return len(a) == 0
}
`,

	structTypeSlice: `
type {{.HelperType}} {{.RawType}}

func (s *{{.HelperType}}) UnmarshalJSONArray(dec *gojay.Decoder) error {
	var value = {{.ComponentInit}}
	if err := dec.Object({{.ComponentPointerModifier}}value); err != nil {
		return err
	}
	*s = append(*s, value)
	return nil
}

func (s {{.HelperType}})  MarshalJSONArray(enc *gojay.Encoder) {
	for i  := range s {
		enc.Object({{.ComponentPointerModifier}}s[i])
	}
}

func (s {{.HelperType}})  IsNil() bool {
	return len(s) == 0
}
`,
	typeSlice: `
type {{.HelperType}} {{.RawType}}

func (s *{{.HelperType}}) UnmarshalJSONArray(dec *gojay.Decoder) error {
	var value = {{.ComponentInit}}
	if err := dec.{{.GojayMethod}}({{.ComponentPointerModifier}}value); err != nil {
		return err
	}
	*s = append(*s, value)
	return nil
}

func (s {{.HelperType}})  MarshalJSONArray(enc *gojay.Encoder) {
	for i  := range s {
		enc.{{.GojayMethod}}({{.ComponentPointerModifier}}s[i])
	}
}

func (s {{.HelperType}})  IsNil() bool {
	return len(s) == 0
}
`,
	timeSlice: `
type {{.HelperType}} {{.RawType}}

func (s *{{.HelperType}}) UnmarshalJSONArray(dec *gojay.Decoder) error {
	var value = {{.ComponentInit}}
	if err := dec.Time({{.ComponentPointerModifier}}value, {{.TimeLayout}}); err != nil {
		return err
	}
	*s = append(*s, value)
	return nil
}

func (s {{.HelperType}})  MarshalJSONArray(enc *gojay.Encoder) {
	for i  := range s {
		enc.Time({{.ComponentPointerModifier}}s[i], {{.TimeLayout}})
	}
}

func (s {{.HelperType}})  IsNil() bool {
	return len(s) == 0
}
`,
	resetStruct: `
// Reset reset fields 
func ({{.Receiver}}) Reset()  {
{{.Reset}}
}
`,

	poolVar: `var {{.PoolName}} *sync.Pool`,
	poolInit: `	{{.PoolName}} = &sync.Pool {
		New: func()interface{} {
			return &{{.Type}}{}
		},
	}`,
	embeddedStructInit: `if {{.Accessor}} == nil { 
		{{.Accessor}} = {{.Init}}
	}`,
}

func expandTemplate(namespace string, dictionary map[int]string, key int, data interface{}) (string, error) {
	var id = fmt.Sprintf("%v_%v", namespace, key)
	textTemplate, ok := dictionary[key]
	if !ok {
		return "", fmt.Errorf("failed to lookup template for %v.%v", namespace, key)
	}
	temlate, err := template.New(id).Parse(textTemplate)
	if err != nil {
		return "", fmt.Errorf("fiailed to parse template %v %v, due to %v", namespace, key, err)
	}
	writer := new(bytes.Buffer)
	err = temlate.Execute(writer, data)
	return writer.String(), err
}

func expandFieldTemplate(key int, data interface{}) (string, error) {
	return expandTemplate("fieldTemplate", fieldTemplate, key, data)
}

func expandBlockTemplate(key int, data interface{}) (string, error) {
	return expandTemplate("blockTemplate", blockTemplate, key, data)
}
