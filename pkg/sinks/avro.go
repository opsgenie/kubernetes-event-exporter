package sinks

// This provides avro encoding using the goavro package.
// Encoding is a simple variation of the Single Object Encoding
// defined in the avro spec:
// https://avro.apache.org/docs/current/spec.html#single_object_encoding
//
// This schemaid is encoded in the leading 17 bytes of the payload
// where the first byte is \0 and the next 16 bytes are the schemaID string

import (
	"encoding/hex"
	"fmt"

	goavro "github.com/linkedin/goavro/v2"
)

// Avro is the config structure to enable avro
// encoding on-demand for the kafka sink
//
// the schemaID is expected to be a 32 char string like an md5hash
//
// the schema must be a legit arvo schema. If the schema does not compile
// then you'll get an error in the log at the time of kafka sink creation
//
// if the incoming event can't be decoded into the given avro schema
// then you'll get an error message in the log and the event will not be
// forwarded to kafka
//

type Avro struct {
	SchemaID string `yaml:"schemaID"`
	Schema   string `yaml:"schema"`
	codec    *goavro.Codec
}

func (a Avro) encode(textual []byte) ([]byte, error) {

	var err error
	dst, err := hex.DecodeString(a.SchemaID)
	if err != nil {
		fmt.Println(string(textual))
		panic(err)
	}

	// make the header
	p := []byte{}
	// leading null byte
	p = append(p, byte(0))
	// shemaid into the next 16 bytes
	p = append(p, dst...)

	// encode the event into avro with the schemid header
	avroNative, _, err := a.codec.NativeFromTextual(textual)
	if err != nil {
		return []byte{}, err
	}

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
