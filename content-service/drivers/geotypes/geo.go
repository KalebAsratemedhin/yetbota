package geotypes

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type NullPoint struct {
	Point *geom.Point
	Valid bool
}

func (n *NullPoint) Scan(value any) error {
	if value == nil {
		n.Point, n.Valid = nil, false
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("NullPoint: expected []byte, got %T", value)
	}
	// lib/pq returns geography as hex-encoded EWKB text
	if isHexBytes(b) {
		decoded, err := hex.DecodeString(string(b))
		if err == nil {
			b = decoded
		}
	}
	g, err := ewkb.Unmarshal(b)
	if err != nil {
		return err
	}
	pt, ok := g.(*geom.Point)
	if !ok {
		return fmt.Errorf("NullPoint: expected *geom.Point, got %T", g)
	}
	n.Point, n.Valid = pt, true
	return nil
}

func isHexBytes(b []byte) bool {
	if len(b) == 0 || len(b)%2 != 0 {
		return false
	}
	for _, c := range b {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (n NullPoint) Value() (driver.Value, error) {
	if !n.Valid || n.Point == nil {
		return nil, nil
	}
	b, err := ewkb.Marshal(n.Point, binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	return hex.EncodeToString(b), nil
}
