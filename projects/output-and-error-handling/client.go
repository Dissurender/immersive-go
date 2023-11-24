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
		time.Sleep(time.Second * 1)
		resp, err := client.Get("http://localhost:8080")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error occured while making request: %s", err)
			fmt.Fprintln(os.Stderr, "\nRetrying fetch..")
			i--
			continue
		}

		e := handleResp(resp)
		if e != nil {
			fmt.Fprintln(os.Stderr, e.Error())
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
		// fmt.Println("Server response ->", string(body))

	case 429:
		retryAfter := res.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {

				if seconds > 5 {
					return fmt.Errorf("Server too busy, please try again later")
				}

				fmt.Fprintf(os.Stderr, "Server busy, waiting for %d seconds.\n", seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
				fmt.Println("retrying...")
			} else if retryTime, err := http.ParseTime(retryAfter); err == nil {

				wait := time.Until(retryTime)
				if wait > 5 {
					return fmt.Errorf("Server too busy, please try again later")
				}

				if wait > 0 {
					fmt.Fprintf(os.Stderr, "Server busy, waiting until %v to retry.", retryTime)
					time.Sleep(wait)
					fmt.Println("retrying...")
				}

			} else {
				return fmt.Errorf("Retry-After value not valid")
			}
		} else {
			return fmt.Errorf("No Retry-After property found")
		}

	default:
		return fmt.Errorf("Unexpected Response Code: %d", res.StatusCode)
	}

	return nil
}
