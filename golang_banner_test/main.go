package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func main() {
	println("Test 1:GET 200")
	
	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set("token", "IGOTTHEPOWER!")
	data := map[string]interface{}{
		"key": "value",
	}

	if err != nil {
		fmt.Printf("Test 1 failed:%s", err.Error())
		return
	}

	client := http.DefaultClient
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Test 1 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Test 1 failed: expected code %d, got %d", http.StatusOK, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Test 1 failed:%s", err.Error())
		return
	}

	var jsonBody map[string]interface{}
	err = json.Unmarshal(body, &jsonBody)

	if err != nil {
		fmt.Printf("Test 1 failed:%s", err.Error())
		return
	}

	if !reflect.DeepEqual(data, jsonBody) {
		fmt.Printf("Test 1 failed: expected %v body, got %v", data, jsonBody)
		return
	}

	println("Test 1:PASS")

	println("Test 2:GET 404")

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=1&feature_id=2", nil)
	req.Header.Set("token", "IGOTTHEPOWER!")

	if err != nil {
		fmt.Printf("Test 2 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 2 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		fmt.Printf("Test 2 failed: expected code %d, got %d", http.StatusNotFound, resp.StatusCode)
		return
	}

	println("Test 2:PASS")

	println("Test 3:GET 403")

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set("token", "IMACREEP")

	if err != nil {
		fmt.Printf("Test 3 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 3 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		fmt.Printf("Test 3 failed: expected code %d, got %d", http.StatusForbidden, resp.StatusCode)
		return
	}

	println("Test 3:PASS")

	println("Test 4:GET 401")

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set("token", "SCAMMER")

	if err != nil {
		fmt.Printf("Test 4 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 4 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		fmt.Printf("Test 4 failed: expected code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		return
	}

	println("Test 4:PASS")

	println("Test 5:GET 400")

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?feature_id=2", nil)
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message := jsonBody["message"]
	error_message := "Invalid format for parameter tag_id: query parameter 'tag_id' is required"

	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3", nil)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message = jsonBody["message"]
	error_message = "Invalid format for parameter feature_id: query parameter 'feature_id' is required"

	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message = jsonBody["message"]
	error_message = "No token was provided"

	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
		return
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=true&feature_id=2", nil)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message = jsonBody["message"]
	error_message = "Invalid format for parameter tag_id: error binding string parameter: strconv.ParseInt: parsing \"true\": invalid syntax"
	
	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
		return
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=2&feature_id=true", nil)
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}
	
	resp, err = client.Do(req)
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}
	
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message = jsonBody["message"]
	error_message = "Invalid format for parameter feature_id: error binding string parameter: strconv.ParseInt: parsing \"true\": invalid syntax"
	
	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
		return
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:8080/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set("token", "")
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}
	
	resp, err = client.Do(req)
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}
	
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("Test 5 failed: expected code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		return
	}

	body, err = io.ReadAll(resp.Body)
	
	if err != nil {
		fmt.Printf("Test 5 failed:%s", err.Error())
		return
	}

	json.Unmarshal(body, &jsonBody)
	message = jsonBody["message"]
	error_message = "Invalid format for parameter token: parameter 'token' is empty, can't bind its value"
	
	if message != error_message {
		fmt.Printf("Test 5 failed: expected message %s, got %s", error_message, message)
		return
	}
	
	println("Test 5:PASS")
	
	println("ALL TESTS HAVE BEEN PASSED!")
}
