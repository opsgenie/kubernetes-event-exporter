package sinks

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	goavro "github.com/linkedin/goavro/v2"
)

// Avro is the config structure to enable avro
// encoding on-demand for the kafka sink
//
// the schemaID is expected to be a 32 char string like an md5hash
//
// the schema must be a legit arvo schema. If the schame does not compile
// then you'll get an error at the time of kafka sink creation
//
// if the incoming event can't be decoded into the given avro schema
// then you'll get an error in the log and the event will not be
// forwarded to kafka
//

type Avro struct {
	SchemaID string `yaml:"schemaID"`
	Schema   string `yaml:"schema"`
	codec    *goavro.Codec
}

func (a Avro) encode(textual []byte) ([]byte, error) {

	// Convert textual Avro data (in Avro JSON format) to native Go form
	// probably don't need this step
	// since it WAS in native Go form in the caller
	// before they converted it to json and sent it here
	// probably just use interface{} which is the native Go form
	// native, _, err := a.codec.NativeFromTextual(dat)
	// if err != nil {
	// log.Error().Err(err).Str("event", string(dat)).Msg("avro schema mismatch")
	// return []byte{}, err
	// }
	// var dat map[string]interface{}
	// if err := json.Unmarshal(byt, &dat); err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dat)

	// schemaIDs look like this
	// c4b52aaf22429c7f9eb8c30270bc1795
	//
	// they are 32 character strings
	// the protocol expects this to be squashed into the 16 bytes
	// the length is checked when the config is loaded
	//
	// squash the schemaID into 16 bytes
	// dst := make([]byte, 16)
	// _, err := squash(dst, a.SchemaID)
	// if err != nil {
	// 	log.Error().Err(err).Msg(string(byt))
	// 	return []byte{}, err
	// }

	// create the proper header
	// which is one null byte
	// followed by the 16 bytes for the ID
	// p := []byte{}
	// p = append(p, byte(0))
	// p = append(p, dst...)

	// Convert native Go form to binary Avro data
	// return a.codec.BinaryFromNative(p, byt)

	// StandardJsonToAvroNative

	// header, avroNative, err := StandardJsonToAvroNative(a.SchemaID, textual)
	// if err != nil {
	// 	fmt.Println(textual)
	// 	panic(err)
	// }

	var err error
	dst, err := hex.DecodeString(a.SchemaID)
	if err != nil {
		fmt.Println(string(textual))
		panic(err)
	}

	p := []byte{}
	p = append(p, byte(0))
	p = append(p, dst...)
	// avrobin, err := a.codec.BinaryFromNative(nil, avroNative)
	// if err != nil {
	// 	fmt.Println(textual)
	// 	panic(err)
	// }

	// for _, buf := range [][]byte{header, avrobin} {
	// 	err = bin.Write(binOut.file, bin.LittleEndian, buf)
	// 	if err != nil {
	// 		fmt.Println(dat)
	// 		panic(err)
	// 	}
	// }

	// this is the new way to go 9/22/21
	avroNative, _, err := a.codec.NativeFromTextual(textual)
	if err != nil {
		return []byte{}, err
	}

	// p = append(p, avroNative...)

	return a.codec.BinaryFromNative(p, avroNative)

}

// NewAvroEncoder creates an encoder which will be used
// to avro encode all events prior to sending to kafka
//
// Its only used by the kafka sink
func NewAvroEncoder(schemaID, schema string) (KafkaEncoder, error) {

	codec, err := goavro.NewCodecForStandardJSON(schema)
	if err != nil {
		return Avro{}, err
	}

	if len(schemaID) != 32 {
		return Avro{}, fmt.Errorf("Avro encoding requires a 32 character schemaID:schema id:%s:", schemaID)
	}

	return Avro{SchemaID: schemaID, codec: codec}, nil
}

// this squashes harder than encode/hex Decode
// by stepping through the input string
// one character at a time, rather than
// one byte at a time
// stuffing two characters into each byte
// of the output byte slice
//
// otherwise this and the function
// fromHexChar are stolen directly from
// encoding/hex
//
// this code is called for every message
// i prefer to leave out the safety nets
// like checking the inputs
// the length is checked when the config is loaded
func squash(dst []byte, src string) (int, error) {
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		a, ok := fromHexChar(src[j-1])
		if !ok {
			return i, hex.InvalidByteError(src[j-1])
		}
		b, ok := fromHexChar(src[j])
		if !ok {
			return i, hex.InvalidByteError(src[j])
		}
		dst[i] = (a << 4) | b
		i++
	}
	if len(src)%2 == 1 {
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if _, ok := fromHexChar(src[j-1]); !ok {
			return i, hex.InvalidByteError(src[j-1])
		}
		return i, hex.ErrLength
	}
	return i, nil
}
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
func StandardJsonToAvroNative(schemaid string, textual []byte) ([]byte, map[string]interface{}, error) {

	//get it into native via encoding json
	// don't use TextualToNative, since it assumes the incoming
	// json is produced per the avro json encoding spec
	// https://avro.apache.org/docs/current/spec.html#json_encoding
	// we don't have that kind of json
	// our comes from kubernetes api and contains union types
	// that are not per that spec and thus won't
	// parse into known avro union types per an avro schema
	var native map[string]interface{}
	if err := json.Unmarshal(textual, &native); err != nil {
		fmt.Println(textual)
		panic(err)
	}

	// t, _, err := codec.NativeFromTextual([]byte("null"))

	// handle the 3 awkward time fields which are avro unions
	// https://github.com/kubernetes/kubernetes/issues/90482#issuecomment-619440160
	// They say "clients must tolerate events without those timestamps set."
	//
	// right now nulls come in as bare null
	// which is per the avro spec so they will be OK
	// https://avro.apache.org/docs/current/spec.html#json_encoding
	//
	// but strings will need to look like this from the spec:
	// the string "a" as {"string": "a"}
	// so rewrite the strings that come in on these unions

	// get this list by scanning the struct for anything where "type" is a list
	// while there get the defaults
	// initially do the top-level but it should be done recursively
	// this can be done once at the start when the codec is compiled
	unions := []string{"eventTime", "firstTimestamp", "lastTimestamp"}
	for _, f := range unions {
		switch v := native[f].(type) {
		case string:
			native[f] = map[string]interface{}{"string": v}
		}
	}

	var err error
	dst, err := hex.DecodeString(schemaid)
	if err != nil {
		fmt.Println(string(textual))
		panic(err)
	}

	p := []byte{}
	p = append(p, byte(0))
	p = append(p, dst...)

	// Convert native Go form to binary Avro data
	// outbin, err := codec.BinaryFromNative(nil, native)
	// return p, outbin, err
	return p, native, nil
}
