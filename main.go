package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "msgpack/msgpack"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalCh
		fmt.Printf("%s\n", sig)
		os.Exit(0)
	}()

	encodeTest()
	decodeTest()
	printInfo()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		switch text {
		case "encode":
			encodeInput()
		case "decode":
			decodeInput()
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unsupported type: ", text)
		}
		printInfo()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input: ", err)
	}
}

func encodeTest() {
	fmt.Println("Encode Example:")
	datas := []map[string]interface{}{
		{"N": 0},
		{"N": 0.1},
		{"N": 0, "M": false},
		{"N": []int{0, 1}, "M": false},
		{"N": map[string]interface{}{"M": "0"}},
		{"N": map[string]interface{}{"M": []int{0, 1}}},
		{"N": map[string]string{"0": "123"}},
	}
	for i, data := range datas {
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshaling data:", err)
			return
		}
		fmt.Print(i+1, ".", string(jsonData))

		b, err := Marshal(data)
		if err != nil {
			fmt.Println("Msgpack encoding error: ", err)
			return
		}

		fmt.Print(" ", hex.EncodeToString(b))
		fmt.Print(" ")
		for i := 0; i < len(b); i++ {
			fmt.Printf("%02x ", b[i])
		}
		fmt.Println("")
	}
}

func encodeInput() {
	var data map[string]interface{}

	fmt.Print("Enter JSON format(support bool/int/float64/map/slice/string): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if err := json.Unmarshal([]byte(input), &data); err != nil {
		fmt.Println("Invalid JSON format: ", err)
		return
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Invalid pretty JSON fotmat: ", err)
	}

	fmt.Println(string(prettyJSON))
	b, err := Marshal(data)
	if err != nil {
		fmt.Println("Msgpack encoding error: ", err)
		return
	}

	fmt.Println(hex.EncodeToString(b))
	for i := 0; i < len(b); i++ {
		fmt.Printf("%02x ", b[i])
	}
}

func decodeTest() {
	fmt.Println("\nDecode Example:")
	datas := [][]byte{
		{0x81, 0xa1, 0x4e, 0xcb, 0x3f, 0xb9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}, // {"N": 0.1}
		{0x81, 0xa1, 0x4d, 0xc2},                                     // {"M":false}
		{0x82, 0xa1, 0x4e, 0x00, 0xa1, 0x4d, 0xc2},                   // {"N": 0, "M": false}
		{0x82, 0xA1, 0x4E, 0x92, 0x00, 0x01, 0xA1, 0x4D, 0xC2},       //{"N": []int{0, 1}, "M": false},
		{0x81, 0xa1, 0x4e, 0x81, 0xa1, 0x4d, 0xa1, 0x30},             //{"N": {"M": "0"}},
		{0x81, 0xA1, 0x4D, 0x92, 0x00, 0x01},                         // {"N": map[string]interface{}{"M": []int{0, 1}}},
		{0x81, 0xA1, 0x4E, 0x81, 0xA1, 0x30, 0xA3, 0x31, 0x32, 0x33}, // {"N": s{"0": "123"}},
	}
	for i, data := range datas {
		fmt.Print(i+1, ".", hex.EncodeToString(data))

		var out map[string]interface{}
		if err := Unmarshal(data, &out); err != nil {
			fmt.Println(" MessagePack encoding error: ", err)
			return
		}

		jsonData, err := json.Marshal(out)
		if err != nil {
			fmt.Println(" Error:", err)
			return
		}
		fmt.Print(" ", string(jsonData))
		fmt.Println("")
	}
}

func decodeInput() {
	fmt.Print("Enter Msgpack Hex String Format: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	inputByte, err := hex.DecodeString(input)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var data map[string]interface{}
	if err := Unmarshal(inputByte, &data); err != nil {
		fmt.Println("Unmarshal error:", err)
		return
	}

	decodeData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(decodeData))
}

func printInfo() {
	fmt.Println("\n\n>>>")
	fmt.Println("Support the following function:")
	fmt.Println("encode | encode JSON to MessagePack format")
	fmt.Println("decode | decode MessagePack to JSON format")
	fmt.Println("exit   | stop program")
	fmt.Println("Ctrl+C | stop program")
	fmt.Print("Input the function: ")
}
