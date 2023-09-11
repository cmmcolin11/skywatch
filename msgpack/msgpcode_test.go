package msgpack_test

import (
    . "app/msgpack"
    "encoding/hex"
    "encoding/json"
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
        "Test float64",
        map[string]interface{}{"M": 0.1},
        "Float64 not support",
    },
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
