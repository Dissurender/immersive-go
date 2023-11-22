package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < 3; i++ {
		resp, err := client.Get("http://localhost:8080")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Status Code: %d", resp.StatusCode)
			fmt.Fprintf(os.Stderr, "Error occured while making request: %s", err)
			os.Exit(1)
		}

		retry := handleResp(resp)
		if retry != nil {
		}
	}

}

func handleResp(res *http.Response) error {
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	switch res.StatusCode {

	case 200:
		bodyBytes, _ := io.ReadAll(res.Body)
		body := string(bodyBytes)
		fmt.Println(body)
		fmt.Println("Server response ->", string(body))

	case 429:
		retryAfter := res.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				fmt.Fprintf(os.Stderr, "Server busy, waiting for %d seconds.", seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
				fmt.Println("retrying...")
			} else if retryTime, err := http.ParseTime(retryAfter); err == nil {

				wait := time.Until(retryTime)
				if wait > 0 {
					fmt.Fprintf(os.Stderr, "Server busy, waiting until %v to retry.", retryTime)
					time.Sleep(wait)
					fmt.Println("retrying...")
				}
			} else {
				fmt.Fprintln(os.Stderr, "Retry-After value not valid.")
				return fmt.Errorf("Retry-After value not valid.")

			}
		} else {
			fmt.Fprintln(os.Stderr, "No Retry-After property found.")
			return fmt.Errorf("No Retry-After property found.")
		}

	default:
		fmt.Fprintf(os.Stderr, "Unexpected Response Code: %d", res.StatusCode)
	}

	return nil
}
