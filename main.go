package main

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const baseUrl = "https://www.googleapis.com/storage/v1/b/chromium-browser-snapshots/o/Win_x64%2F"

func main() {
	os.Exit(m())
}

func m() int {
	//get LAST_CHANGE metadata
	resp, err := http.Get(baseUrl + "LAST_CHANGE")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error making LAST_CHANGE metadata request:", err)
		return 1
	}
	r := bytes.Buffer{}
	n, err := io.Copy(&r, resp.Body)
	if err != nil || n != resp.ContentLength {
		_, _ = fmt.Fprintln(os.Stderr, "Error downloading metadata reponse body:", err)
		return 1
	}
	err = resp.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error closing metadata request response:", err)
		return 1
	}

	//parse LAST_CHANGE metadata to extract download url for it
	ml, err := jsonparser.GetString(r.Bytes(), "mediaLink")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error parsing LAST_CHANGE metadata json response:", err)
		return 1
	}

	//get LAST_CHANGE file
	resp, err = http.Get(ml)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error downloading LAST_CHANGE:", err)
		return 1
	}
	r = bytes.Buffer{}
	n, err = io.Copy(&r, resp.Body)
	if err != nil || n != resp.ContentLength {
		_, _ = fmt.Fprintln(os.Stderr, "Error copying LAST_CHANGE response body:", err)
		return 1
	}

	vs := r.String()
	fmt.Println("LAST_CHANGE:", vs)

	//get mini installer metadata
	resp, err = http.Get(baseUrl + vs + "%2Fmini_installer.exe")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error making installer metadata request:", err.Error())
		return 1
	}
	r = bytes.Buffer{}
	n, err = io.Copy(&r, resp.Body)
	if err != nil || n != resp.ContentLength {
		_, _ = fmt.Fprintln(os.Stderr, "Error downloading installer metadata:", err)
		return 1
	}
	err = resp.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error closing installer metadata response:", err)
		return 1
	}

	//parse mini installer download url
	ml, err = jsonparser.GetString(r.Bytes(), "mediaLink")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error parsing installer metadata json response:", err)
		return 1
	}

	//download mini installer
	resp, err = http.Get(ml)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error making installer request:", err.Error())
		return 1
	}
	f, err := os.Create("mini_installer_" + vs + ".exe") //installer download file
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error creating installer file:", err)
		return 1
	}
	fmt.Print("Downloading installer...")
	n, err = io.Copy(f, resp.Body)
	if err != nil || n != resp.ContentLength {
		_, _ = fmt.Fprintln(os.Stderr, "Error downloading installer:", err)
		return 1
	}
	err = resp.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error closing installer response:", err)
		return 1
	}
	err = f.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error closing installer file:", err)
		return 1
	}
	fmt.Println("done")

	//run installer
	if runtime.GOOS == "windows" {
		cmd := exec.Command("mini_installer" + r.String() + ".exe")
		err = cmd.Run()
		if err != nil {

		}
		err = cmd.Wait()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error waiting for installer to finish executing:", err)
			return 1
		}
	} else {
		fmt.Println("non-windows os, not running installer.")
	}

	return 0
}
