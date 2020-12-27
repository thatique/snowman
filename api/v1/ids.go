package v1

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/jsonpb"
)

// ID returned in GRPC
type ID uint64

// NewIDFromString create ID from string
func NewIDFromString(s string) (ID, error) {
	if len(s) > 16 {
		return ID(0), fmt.Errorf("ID cannot be longer than 16 hex character: %s", s)
	}
	id, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return ID(0), err
	}

	return ID(id), nil
}

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

// MarshalJSON converts id into a string enclosed in quotes.
func (id ID) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%s"`, id.String())
	return []byte(s), nil
}

// UnmarshalJSON inflates id from base64 string, possibly enclosed in quotes.
func (id *ID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if l := len(str); l > 2 && str[0] == '"' && str[l-1] == '"' {
		str = str[1 : l-1]
	}
	nid, err := NewIDFromString(str)
	if err != nil {
		return err
	}

	*id = nid
	return nil
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
