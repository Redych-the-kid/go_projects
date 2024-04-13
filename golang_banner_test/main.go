package main


import (
	"io"
    "fmt"
    "net/http"
	"encoding/json"
	"reflect"
)

func main() {
	println("Test 1:GET 200")
	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set("token", "IGOTTHEPOWER!")
	data := map[string]interface{}{
		"key": "value",
	}
	if err != nil{
		fmt.Printf("Test failed:%s", err.Error())
		return
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
    	fmt.Printf("Test failed:%s", err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil{
		fmt.Printf("Test failed:%s", err.Error())
		return
	}
	var jsonBody map[string]interface{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil{
		fmt.Printf("Test failed:%s", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK{
		fmt.Printf("Test failed: expected code %d, got %d", resp.StatusCode, http.StatusOK)
		return
	}
	if !reflect.DeepEqual(data, jsonBody){
		fmt.Printf("Test failed: expected %v body, got %v", data, jsonBody)
		return
	}
	println("Test 1:PASS")
}