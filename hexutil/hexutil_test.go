package hexutil

import (
	"bytes"
	"math/big"
	"testing"
)

type marshalTest struct {
	input interface{}
	want  string
}

type unmarshalTest struct {
	input        string
	want         interface{}
	wantErr      error // if set, decoding must fail on any platform
	wantErr32bit error // if set, decoding must fail on 32bit platforms (used for Uint tests)
}

var (
	encodeBytesTests = []marshalTest{
		{[]byte{}, ""},
		{[]byte{0}, "00"},
		{[]byte{0, 0, 1, 2}, "00000102"},
	}

	encodeBigTests = []marshalTest{
		{referenceBig("0"), "0"},
		{referenceBig("1"), "1"},
		{referenceBig("ff"), "ff"},
		{referenceBig("112233445566778899aabbccddeeff"), "112233445566778899aabbccddeeff"},
		{referenceBig("80a7f2c1bcc396c00"), "80a7f2c1bcc396c00"},
		{referenceBig("-80a7f2c1bcc396c00"), "-80a7f2c1bcc396c00"},
	}

	encodeUint64Tests = []marshalTest{
		{uint64(0), "0"},
		{uint64(1), "1"},
		{uint64(0xff), "ff"},
		{uint64(0x1122334455667788), "1122334455667788"},
	}

	encodeUintTests = []marshalTest{
		{uint(0), "0"},
		{uint(1), "1"},
		{uint(0xff), "ff"},
		{uint(0x11223344), "11223344"},
	}

	decodeBytesTests = []unmarshalTest{
		// invalid
		{input: `0x`, wantErr: ErrEmptyString},
		{input: `0x0`, wantErr: ErrOddLength},
		{input: `0x023`, wantErr: ErrOddLength},
		{input: `0xxx`, wantErr: ErrSyntax},
		{input: `0x01zz01`, wantErr: ErrSyntax},
		// invalid without 0x prefix
		{input: ``, wantErr: ErrEmptyString},
		{input: `0`, wantErr: ErrOddLength},
		{input: `023`, wantErr: ErrOddLength},
		{input: `xx`, wantErr: ErrSyntax},
		{input: `01zz01`, wantErr: ErrSyntax},
		// valid
		{input: `0x02`, want: []byte{0x02}},
		{input: `0X02`, want: []byte{0x02}},
		{input: `0xffffffffff`, want: []byte{0xff, 0xff, 0xff, 0xff, 0xff}},
		{
			input: `0xffffffffffffffffffffffffffffffffffff`,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		// valid without 0x prefix
		{input: `02`, want: []byte{0x02}},
		{input: `ffffffffff`, want: []byte{0xff, 0xff, 0xff, 0xff, 0xff}},
		{
			input: `ffffffffffffffffffffffffffffffffffff`,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
	}

	decodeBigTests = []unmarshalTest{
		// invalid
		{input: ``, wantErr: ErrEmptyString},
		{input: `01`, wantErr: ErrLeadingZero},
		{input: `x`, wantErr: ErrSyntax},
		{input: `1zz01`, wantErr: ErrSyntax},
		{
			input:   `0x10000000000000000000000000000000000000000000000000000000000000000`,
			wantErr: ErrBig256Range,
		},
		// invalid with 0x prefix
		{input: `0x`, wantErr: ErrEmptyNumber},
		{input: `0x01`, wantErr: ErrLeadingZero},
		{input: `0xx`, wantErr: ErrSyntax},
		{input: `0x1zz01`, wantErr: ErrSyntax},
		{
			input:   `0x10000000000000000000000000000000000000000000000000000000000000000`,
			wantErr: ErrBig256Range,
		},
		// valid
		{input: `0`, want: big.NewInt(0)},
		{input: `2`, want: big.NewInt(0x2)},
		{input: `2F2`, want: big.NewInt(0x2f2)},
		{input: `1122aaff`, want: big.NewInt(0x1122aaff)},
		{input: `bBb`, want: big.NewInt(0xbbb)},
		{input: `fffffffff`, want: big.NewInt(0xfffffffff)},
		{
			input: `112233445566778899aabbccddeeff`,
			want:  referenceBig("112233445566778899aabbccddeeff"),
		},
		{
			input: `ffffffffffffffffffffffffffffffffffff`,
			want:  referenceBig("ffffffffffffffffffffffffffffffffffff"),
		},
		{
			input: `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
			want:  referenceBig("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		},
		// valid with 0x prefix
		{input: `0x0`, want: big.NewInt(0)},
		{input: `0x2`, want: big.NewInt(0x2)},
		{input: `0x2F2`, want: big.NewInt(0x2f2)},
		{input: `0X2F2`, want: big.NewInt(0x2f2)},
		{input: `0x1122aaff`, want: big.NewInt(0x1122aaff)},
		{input: `0xbBb`, want: big.NewInt(0xbbb)},
		{input: `0xfffffffff`, want: big.NewInt(0xfffffffff)},
		{
			input: `0x112233445566778899aabbccddeeff`,
			want:  referenceBig("112233445566778899aabbccddeeff"),
		},
		{
			input: `0xffffffffffffffffffffffffffffffffffff`,
			want:  referenceBig("ffffffffffffffffffffffffffffffffffff"),
		},
		{
			input: `0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
			want:  referenceBig("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		},
	}

	decodeUint64Tests = []unmarshalTest{
		// invalid
		{input: `0x`, wantErr: ErrEmptyNumber},
		{input: `0x01`, wantErr: ErrLeadingZero},
		{input: `0xfffffffffffffffff`, wantErr: ErrUint64Range},
		{input: `0xx`, wantErr: ErrSyntax},
		{input: `0x1zz01`, wantErr: ErrSyntax},
		// invalid with 0x prefix
		{input: `0x`, wantErr: ErrEmptyNumber},
		{input: `0x01`, wantErr: ErrLeadingZero},
		{input: `0xfffffffffffffffff`, wantErr: ErrUint64Range},
		{input: `0xx`, wantErr: ErrSyntax},
		{input: `0x1zz01`, wantErr: ErrSyntax},
		// valid
		{input: `0`, want: uint64(0)},
		{input: `2`, want: uint64(0x2)},
		{input: `2F2`, want: uint64(0x2f2)},
		{input: `2F2`, want: uint64(0x2f2)},
		{input: `1122aaff`, want: uint64(0x1122aaff)},
		{input: `bbb`, want: uint64(0xbbb)},
		{input: `ffffffffffffffff`, want: uint64(0xffffffffffffffff)},
		// valid with 0x prefix
		{input: `0x0`, want: uint64(0)},
		{input: `0x2`, want: uint64(0x2)},
		{input: `0x2F2`, want: uint64(0x2f2)},
		{input: `0X2F2`, want: uint64(0x2f2)},
		{input: `0x1122aaff`, want: uint64(0x1122aaff)},
		{input: `0xbbb`, want: uint64(0xbbb)},
		{input: `0xffffffffffffffff`, want: uint64(0xffffffffffffffff)},
	}
)

func TestEncode(t *testing.T) {
	for _, test := range encodeBytesTests {
		enc := Encode(test.input.([]byte))
		if enc != test.want {
			t.Errorf("input %x: wrong encoding %s", test.input, enc)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, test := range decodeBytesTests {
		dec, err := Decode(test.input)
		if !checkError(t, test.input, err, test.wantErr) {
			continue
		}
		if !bytes.Equal(test.want.([]byte), dec) {
			t.Errorf("input %s: value mismatch: got %x, want %x", test.input, dec, test.want)
			continue
		}
	}
}

func TestEncodeBig(t *testing.T) {
	for _, test := range encodeBigTests {
		enc := EncodeBig(test.input.(*big.Int))
		if enc != test.want {
			t.Errorf("input %x: wrong encoding %s", test.input, enc)
		}
	}
}

func TestDecodeBig(t *testing.T) {
	for _, test := range decodeBigTests {

		dec, err := DecodeBig(test.input)
		if !checkError(t, test.input, err, test.wantErr) {
			continue
		}
		if dec.Cmp(test.want.(*big.Int)) != 0 {
			t.Errorf("input %s: value mismatch: got %x, want %x", test.input, dec, test.want)
			continue
		}
	}
}

func TestEncodeUint64(t *testing.T) {
	for _, test := range encodeUint64Tests {
		enc := EncodeUint64(test.input.(uint64))
		if enc != test.want {
			t.Errorf("input %x: wrong encoding %s", test.input, enc)
		}
	}
}

func TestDecodeUint64(t *testing.T) {
	for _, test := range decodeUint64Tests {
		dec, err := DecodeUint64(test.input)
		if !checkError(t, test.input, err, test.wantErr) {
			continue
		}
		if dec != test.want.(uint64) {
			t.Errorf("input %s: value mismatch: got %x, want %x", test.input, dec, test.want)
			continue
		}
	}
}
