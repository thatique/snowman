package v1

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
)

// ID returned in GRPC
type ID uint64

// String returns string representation of ID
func (id ID) String() string {
	return fmt.Sprintf("%x", uint64(id))
}

// Size returns the size of this datum in protobuf. It is always 8 bytes.
func (id *ID) Size() int {
	return 8
}

// MarshalTo converts ID into a binary representation. Called by protobuf serialization.
func (id *ID) MarshalTo(data []byte) (n int, err error) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(*id))
	return marshalBytes(data, b[:])
}

// Unmarshal inflates ID from a binary representation. Called by protobuf serialization.
func (id *ID) Unmarshal(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("buffer is too short")
	}
	*id = ID(binary.BigEndian.Uint64(data))
	return nil
}

// MarshalJSON converts span id into a base64 string enclosed in quotes.
// Used by protobuf JSON serialization.
// Example: {1} => "AAAAAAAAAAE=".
func (id ID) MarshalJSON() ([]byte, error) {
	var b [8]byte
	id.MarshalTo(b[:]) // can only error on incorrect buffer size
	v := make([]byte, 12+2)
	base64.StdEncoding.Encode(v[1:13], b[:])
	v[0], v[13] = '"', '"'
	return v, nil
}

// UnmarshalJSON inflates id from base64 string, possibly enclosed in quotes.
// User by protobuf JSON serialization.
//
// There appears to be a bug in gogoproto, as this function is only called for numeric values.
// https://github.com/gogo/protobuf/issues/411#issuecomment-393856837
func (id *ID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if l := len(str); l > 2 && str[0] == '"' && str[l-1] == '"' {
		str = str[1 : l-1]
	}
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return fmt.Errorf("cannot unmarshal ID from string '%s': %v", string(data), err)
	}
	return id.Unmarshal(b)
}

// UnmarshalJSONPB inflates id from base64 string, possibly enclosed in quotes.
// User by protobuf JSON serialization.
//
// TODO: can be removed once this ticket is fixed:
//       https://github.com/gogo/protobuf/issues/411#issuecomment-393856837
func (id *ID) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, b []byte) error {
	return id.UnmarshalJSON(b)
}

func marshalBytes(dst []byte, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		return 0, fmt.Errorf("buffer is too short")
	}
	return copy(dst, src), nil
}