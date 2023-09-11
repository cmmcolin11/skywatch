package msgpack_test

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    . "msgpack/msgpack"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

type encoderTest struct {
    name     string
    in       interface{}
    expected string
}

var encoderTests = []encoderTest{
    {
        "Test bool",
        map[string]interface{}{"M": true},
        "81a14dc3",
    },
    {
        "Test int64",
        map[string]interface{}{"M": 0},
        "81a14d00",
    },
    {
        "Test float64",
        map[string]interface{}{"M": 0.1},
        "81a14dcb3fb999999999999a",
    },
    {
        "Test slice",
        map[string]interface{}{"M": []int{0, 1}},
        "81a14d920001",
    },
    {
        "Test string",
        map[string]interface{}{"M": "0"},
        "81a14da130",
    },
    {
        "Test nested",
        map[string]interface{}{"M": map[string]interface{}{"I": true, "J": 0, "K": []int{0, 1}, "L": "0"}},
        "81a14d84a149c3a14a00a14b920001a14ca130",
    },
}

var encoderNotSupportTests = []encoderTest{
    {
        "Test ptr",
        map[string]interface{}{"M": new(time.Time)},
        "Ptr not support",
    },
}

func TestMarshal(t *testing.T) {
    for i, test := range encoderTests {
        b, _ := Marshal(test.in)
        result := hex.EncodeToString(b)
        require.Equal(t, test.expected, result, "#%d", i)

        testIn, _ := json.Marshal(test.in)
        t.Logf("(%d) %s\nInput: %s\nExpected: %s\nActual:   %s\n", i+1, test.name, testIn, test.expected, result)
    }

    n := len(encoderTests)
    for i, test := range encoderNotSupportTests {
        j := i + n
        testIn, _ := json.Marshal(test.in)
        _, err := Marshal(test.in)
        if err == nil {
            t.Logf("(%d) %s\nInput: %s\nExpected: %s\nActual: Not PASS\n", j+1, test.name, string(testIn), test.expected)
        } else {
            t.Logf("(%d) %s\nInput: %s\nExpected: %s\nActual: PASS\n", j+1, test.name, string(testIn), test.expected)
        }
    }
}

//--------------------------------------------------

type decoderTest struct {
    name     string
    in       []byte
    expected string
}

var decoderTests = []decoderTest{
    {
        "Test bool",
        []byte{0x81, 0xa1, 0x4d, 0xc2},
        "{\"M\":false}",
    },
    {
        "Test int64",
        []byte{0x81, 0xA1, 0x4D, 0x00},
        "{\"M\":0}",
    },
    {
        "Test float64",
        []byte{0x81, 0xA1, 0x4D, 0xCB, 0x3F, 0xB9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9A},
        "{\"M\":0.1}",
    },
    {
        "Test slice",
        []byte{0x81, 0xA1, 0x4D, 0x92, 0x00, 0x01},
        "{\"M\":[0,1]}",
    },
    {
        "Test string",
        []byte{0x81, 0xA1, 0x4D, 0xA1, 0x30},
        "{\"M\":\"0\"}",
    },
    {
        "Test nested",
        []byte{0x84, 0xA1, 0x49, 0xC3, 0xA1, 0x4A, 0x00, 0xA1, 0x4B, 0x92, 0x00, 0x01, 0xA1, 0x4C, 0xA1, 0x30},
        "{\"I\":true,\"J\":0,\"K\":[0,1],\"L\":\"0\"}",
    },
}

func TestUnmarshal(t *testing.T) {
    for i, test := range decoderTests {
        var data map[string]interface{}
        if err := Unmarshal(test.in, &data); err != nil {
            fmt.Println("MessagePack Unmarshal error: ", err)
        }

        result, err := json.Marshal(data)
        if err != nil {
            fmt.Println("JSON Marshal Error:", err)
            return
        }
        require.Equal(t, test.expected, string(result), "#%d", i)
        t.Logf("(%d) %s\nInput: %x\nExpected: %s\nActual:   %s\n", i+1, test.name, test.in, test.expected, string(result))
    }
}
