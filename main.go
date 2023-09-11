package main

import (
	. "msgpack/msgpack"
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	printInfo()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		switch text {
		case "encode":
			encodeInput()
		case "decode":
			fmt.Println("Not supported yet")
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

func printInfo() {
	fmt.Println("\n\n>>>")
	fmt.Println("Support the following function:")
	fmt.Println("encode | encode JSON to MessagePack format")
	fmt.Println("decode | decode MessagePack to JSON format")
	fmt.Println("exit   | stop program")
	fmt.Println("Ctrl+C | stop program")
	fmt.Print("Input the function: ")
}
