package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SERVER_HOST    = "localhost"
	SERVER_PORT    = "9980"
	SERVER_TYPE    = "tcp"
	INDEX_BODY     = "File Download Successfully"
	NOT_FOUND_BODY = "The requested file does not found"
)

func main() {
	fmt.Println("Server Running...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting new connection: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("A new client connected")
		go processClient(connection)
	}
}

func processClient(connection net.Conn) {
	defer connection.Close()

	path, err := readAndDecodeHTTPFirstLine(connection)

	if err != nil {
		fmt.Println("Error decoding first line of request:", err.Error())
		return
	}

	foldername := os.Args[1]
	if foldername == "" {
		fmt.Println("Could not found Arguments in first line ")
		return
	}

	fmt.Printf("Got new 'GET' request for %s\n", path)
	requestHeaders, err := readAndDecodeHTTPHeaders(connection)
	fmt.Println("r   ", requestHeaders)
	host, foundHost := requestHeaders["Host"]
	fmt.Println(host)
	if !foundHost {
		fmt.Println("Could not found 'Host' in request headers")
		return
	}
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	host_minus_port := strings.Split(host, ":")
	hostpath := "http://" + host_minus_port[0] + "/" + foldername + path
	path_minus_file := strings.Split(path, "/")
	pathfile := path_minus_file[1]

	body := ""
	statusCode := 0
	statusMessage := ""
	if path != "" {

		if foldername != "" {

			if exists(pathfile) {
				err := DownloadFile(path, hostpath)
				if err != nil {
					body = NOT_FOUND_BODY
					statusCode = 404
					statusMessage = "Could not write in file correctly"

				} else {
					fmt.Println("Downloaded " + path)
					body = INDEX_BODY
					statusCode = 200
					statusMessage = "OK"
				}
			} else {
				if path == "/" {
					body = " <a href='http://127.0.0.1/new/index.html'>Address does not exist, Click to redirect Index.html</a>"

					statusCode = 302
					statusMessage = "Could not found this file then redirect to index.html"
				} else {
					body = NOT_FOUND_BODY

					statusCode = 404
					statusMessage = "file does not exist"
				}
			}
		}
	}

	responseHeaders := make(map[string]string)
	responseHeaders["Content-Length"] = strconv.Itoa(len(body))
	responseHeaders["Host"] = host
	response := "HTTP/1.1 " + strconv.Itoa(statusCode) + " " + statusMessage + "\r\n" + encodeHTTPHeaders(responseHeaders) + "\r\n" + body
	_, err = connection.Write([]byte(response))
	if err != nil {
		fmt.Println("Could not write response to connection:", err.Error())
	}
	fmt.Println("Sent response with status code:", statusCode)
}

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
func DownloadFile(fileName string, urlstr string) error {

	base := filepath.Base(urlstr)

	resp, err := http.Get(urlstr)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("errorr")
	}
	b, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

	if err := os.WriteFile(base, b, 0644); err != nil {
		panic(err)
	}

	return err
}

func readAndDecodeHTTPFirstLine(connection net.Conn) (string, error) {
	line, err := readLineFromConnection(connection)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not read a new line from connection: %s", err.Error()))
	}
	fmt.Println("line", line)
	partList := strings.Split(line, " ")
	if len(partList) > 3 {
		return "", errors.New("More than 3 parts in first line")
	}
	if len(partList) < 3 {
		return "", errors.New("Less than 3 parts in first line")
	}
	method := partList[0]
	if method != "GET" {
		return "", errors.New(fmt.Sprintf("Unhandled method '%s'", method))
	}
	path := partList[1]
	version := partList[2]
	if version != "HTTP/1.1" {
		return "", errors.New(fmt.Sprintf("Unhandled method: '%s'", version))
	}
	return path, nil
}

func readLineFromConnection(connection net.Conn) (string, error) {
	data := ""
	for {
		// Read one byte per loop iteration since we want to stop reading after reaching "\r\n"
		buffer := make([]byte, 1)
		_, err := connection.Read(buffer)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Could not read new data from connection: '%s'", err.Error()))
		}
		data = data + string(buffer[0])
		if strings.HasSuffix(data, "\r\n") {
			data = data[:len(data)-2] // Remove "\r\n"
			break
		}
	}
	fmt.Printf("Read a new line from connection: '%s'\n", data)
	return data, nil
}

func readAndDecodeHTTPHeaders(connection net.Conn) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		line, err := readLineFromConnection(connection)
		if line == "" {
			break
		}
		if err != nil {
			return headers, err
		}
		linePartList := strings.SplitAfterN(line, ":", 2) // SplitAfterN keeps ":" after parts

		if len(linePartList) < 2 {
			return headers, errors.New(fmt.Sprintf("Less than two parts in header line '%s'", line))
		}

		key := strings.TrimSpace(linePartList[0])
		key = key[:len(key)-1] // remove ":" at the end of key
		value := strings.TrimSpace(linePartList[1])
		fmt.Printf("Decoded new header line with value '%s' and value '%s'\n", key, value)
		headers[key] = value
	}
	return headers, nil
}

func encodeHTTPHeaders(headers map[string]string) string {
	result := ""
	for key, value := range headers {
		result = result + key + ": " + value + "\r\n"
	}
	return result
}
