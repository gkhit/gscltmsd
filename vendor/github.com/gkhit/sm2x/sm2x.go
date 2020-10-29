package sm2x

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
	"sort"
)

// ConvParameters Conversion parameters
type ConvParameters struct {
	XMLEscapeChars       bool // if 'true' forces escaping invalid characters in attribute and element values
	ScientificFloat      int8 // if 'true' output float value in scientific notation, e.g. -1.234456e+78
	GoEmptyElementSyntax bool // if 'true' encode empty element as "<tag></tag>" instead of "<tag/>
	SkipUnknown          bool // if 'true' skip unknown elements, else append '>UNKNOWN'
	AppendHeader         bool // if 'true' append xml header '<?xml version="1.0" encoding="utf-8"?>'
	DefaultRootTag       string
}

var defaultConvParameters = &ConvParameters{
	XMLEscapeChars:       true,
	ScientificFloat:      0,
	GoEmptyElementSyntax: false,
	SkipUnknown:          false,
	AppendHeader:         false,
	DefaultRootTag:       "doc",
}

// DefaultConversionParameters Return default conversion parameters
func DefaultConversionParameters() *ConvParameters {
	p := new(ConvParameters)
	*p = *defaultConvParameters
	return p
}

// Map2XML Convert string map to XML
func Map2XML(m map[string]interface{}, rootTag ...string) ([]byte, error) {
	var err error
	s := new(string)
	p := new(ConvParameters)
	*p = *defaultConvParameters

	if len(m) == 1 && len(rootTag) == 0 {
		for key, value := range m {
			// if it an array, see if all values are map[string]interface{}
			// we force a new root tag if we'll end up with no key:value in the list
			// so: key:[string_val, bool:true] --> <doc><key>string_val</key><bool>true</bool></key></doc>
			switch value.(type) {
			case []interface{}:
				for _, v := range value.([]interface{}) {
					switch v.(type) {
					case map[string]interface{}: // noop
					default: // anything else
						err = mapToXML(s, p.DefaultRootTag, m, p)
						goto done
					}
				}
			}
			err = mapToXML(s, key, value, p)
		}
	} else if len(rootTag) == 1 {
		err = mapToXML(s, rootTag[0], m, p)
	} else {
		err = mapToXML(s, p.DefaultRootTag, m, p)
	}
done:
	if p.AppendHeader {
		*s = "<?xml version=\"1.0\" encoding=\"utf-8\"?>" + *s
	}
	return []byte(*s), err
}

// Map2XMLParameters Convert string map to XML with parameters
func Map2XMLParameters(m map[string]interface{}, params *ConvParameters, rootTag ...string) ([]byte, error) {
	var err error
	s := new(string)
	p := new(ConvParameters)
	*p = *params

	if len(m) == 1 && len(rootTag) == 0 {
		for key, value := range m {
			// if it an array, see if all values are map[string]interface{}
			// we force a new root tag if we'll end up with no key:value in the list
			// so: key:[string_val, bool:true] --> <doc><key>string_val</key><bool>true</bool></key></doc>
			switch value.(type) {
			case []interface{}:
				for _, v := range value.([]interface{}) {
					switch v.(type) {
					case map[string]interface{}: // noop
					default: // anything else
						err = mapToXML(s, p.DefaultRootTag, m, p)
						goto done
					}
				}
			}
			err = mapToXML(s, key, value, p)
		}
	} else if len(rootTag) == 1 {
		err = mapToXML(s, rootTag[0], m, p)
	} else {
		err = mapToXML(s, p.DefaultRootTag, m, p)
	}
done:
	if p.AppendHeader {
		*s = "<?xml version=\"1.0\" encoding=\"utf-8\"?>" + *s
	}
	return []byte(*s), err
}

// order is important - must scan for '&' first
var escapechars = [][2][]byte{
	{[]byte(`&`), []byte(`&amp;`)},
	{[]byte(`<`), []byte(`&lt;`)},
	{[]byte(`>`), []byte(`&gt;`)},
	{[]byte(`"`), []byte(`&quot;`)},
	{[]byte(`'`), []byte(`&apos;`)},
}

func escapeChars(s string) string {
	if len(s) == 0 {
		return s
	}

	b := []byte(s)
	for _, v := range escapechars {
		n := bytes.Count(b, v[0])
		if n == 0 {
			continue
		}
		b = bytes.Replace(b, v[0], v[1], n)
	}
	return string(b)
}

// where the work actually happens
// returns an error if an attribute is not atomic
func mapToXML(s *string, key string, value interface{}, p *ConvParameters) error {
	var endTag bool
	var elen int

	if reflect.ValueOf(value).Kind() == reflect.Map {
		switch value.(type) {
		case map[string]interface{}:
		default:
			val := make(map[string]interface{})
			vv := reflect.ValueOf(value)
			keys := vv.MapKeys()
			for _, k := range keys {
				val[fmt.Sprint(k)] = vv.MapIndex(k).Interface()
			}
			value = val
		}
	}

	switch value.(type) {
	case map[string]interface{}, []byte, string, float64, bool, int, int32, int64, float32:
		*s += `<` + key
	}
	switch value.(type) {
	case map[string]interface{}:
		vv := value.(map[string]interface{})
		lenvv := len(vv)
		// scan out attributes - keys have prepended hyphen, '-'
		attrlist := make([][2]string, len(vv))
		var n int
		var ss string
		for k, v := range vv {
			if k[:1] == "-" {
				switch v.(type) {
				case string:
					ss = v.(string)
					if p.XMLEscapeChars {
						ss = escapeChars(ss)
					}
					attrlist[n][0] = k[1:]
					attrlist[n][1] = ss
				case bool, int, int32, int64:
					attrlist[n][0] = k[1:]
					attrlist[n][1] = fmt.Sprintf("%v", v)
				case float64, float32:
					attrlist[n][0] = k[1:]
					switch p.ScientificFloat {
					case 'e':
						attrlist[n][1] = fmt.Sprintf("%e", v)
					case 'f':
						attrlist[n][1] = fmt.Sprintf("%f", v)
					default:
						attrlist[n][1] = fmt.Sprintf("%v", v)
					}
				case []byte:
					ss = string(v.([]byte))
					if p.XMLEscapeChars {
						ss = escapeChars(ss)
					}
					attrlist[n][0] = k[1:]
					attrlist[n][1] = ss
				default:
					return fmt.Errorf("invalid attribute value for: %s", k)
				}
				n++
			}
		}
		if n > 0 {
			attrlist = attrlist[:n]
			sort.Sort(attrList(attrlist))
			for _, v := range attrlist {
				*s += ` ` + v[0] + `="` + v[1] + `"`
			}
		}

		// only attributes?
		if n == lenvv {
			break
		}
		// simple element? Note: '#text" is an invalid XML tag.
		if v, ok := vv["#text"]; ok && n+1 == lenvv {
			switch v.(type) {
			case string:
				if p.XMLEscapeChars {
					v = escapeChars(v.(string))
				} else {
					v = v.(string)
				}
			case []byte:
				if p.XMLEscapeChars {
					v = escapeChars(string(v.([]byte)))
				}
			}
			*s += ">" + fmt.Sprintf("%v", v)
			endTag = true
			elen = 1
			break
		}
		// close tag with possible attributes
		*s += ">"
		// extract the map k:v pairs and sort on key
		elemlist := make([][2]interface{}, len(vv))
		n = 0
		for k, v := range vv {
			if k[:1] == "-" {
				continue
			}
			elemlist[n][0] = k
			elemlist[n][1] = v
			n++
		}
		elemlist = elemlist[:n]
		sort.Sort(elemList(elemlist))
		var i int
		for _, v := range elemlist {
			i++
			mapToXML(s, v[0].(string), v[1], p)
			i--
		}
		endTag = true
		elen = 1 // we do have some content ...
	case []interface{}:
		for _, v := range value.([]interface{}) {
			mapToXML(s, key, v, p)
		}
		return nil
	case nil:
		// terminate the tag
		*s += "<" + key
		endTag = true
		break
	default: // handle anything - even goofy stuff
		elen = 0
		switch value.(type) {
		case string:
			v := value.(string)
			if p.XMLEscapeChars {
				v = escapeChars(v)
			}
			elen = len(v)
			if elen > 0 {
				*s += ">" + v
			}
		case bool, int, int32, int64:
			v := fmt.Sprintf("%v", value)
			elen = len(v) // always > 0
			*s += ">" + v
		case float64, float32:
			switch p.ScientificFloat {
			case 'e':
				v := fmt.Sprintf("%e", value)
				elen = len(v) // always > 0
				*s += ">" + v
			case 'f':
				v := fmt.Sprintf("%f", value)
				elen = len(v) // always > 0
				*s += ">" + v
			default:
				v := fmt.Sprintf("%v", value)
				elen = len(v) // always > 0
				*s += ">" + v
			}
		case []byte: // NOTE: byte is just an alias for uint8
			// similar to how xml.Marshal handles []byte structure members
			v := string(value.([]byte))
			if p.XMLEscapeChars {
				v = escapeChars(v)
			}
			elen = len(v)
			if elen > 0 {
				*s += ">" + v
			}
		default:
			var v []byte
			var err error
			v, err = xml.Marshal(value)
			if err != nil {
				if !p.SkipUnknown {
					*s += ">UNKNOWN"
				}
			} else {
				elen = len(v)
				if elen > 0 {
					*s += string(v)
				}
			}
		}
		endTag = true
	}
	if endTag {
		switch value.(type) {
		case map[string]interface{}, []byte, string, float64, bool, int, int32, int64, float32, nil:
			if elen > 0 || p.GoEmptyElementSyntax {
				if elen == 0 {
					*s += ">"
				}
				*s += `</` + key + ">"
			} else {
				*s += `/>`
			}
		}
	} else if p.GoEmptyElementSyntax {
		*s += "></" + key + ">"
	} else {
		*s += "/>"
	}

	return nil
}

// ============================ sort interface implementation =================

type attrList [][2]string

func (a attrList) Len() int {
	return len(a)
}

func (a attrList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a attrList) Less(i, j int) bool {
	if a[i][0] > a[j][0] {
		return false
	}
	return true
}

type elemList [][2]interface{}

func (e elemList) Len() int {
	return len(e)
}

func (e elemList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e elemList) Less(i, j int) bool {
	if e[i][0].(string) > e[j][0].(string) {
		return false
	}
	return true
}
