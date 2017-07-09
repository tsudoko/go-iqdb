// Package iqdb provides a client for the iqdb protocol.
package iqdb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	conn net.Conn
}

type Response struct {
	Code    int
	Content string
}

type QueryResult struct {
	ImgID  int
	Score  float64
	Width  int
	Height int
}

type MultiQueryResult struct {
	QueryResult
	DbID int
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.New("Dial: " + err.Error())
	}

	return &Client{conn}, nil
}

func (c *Client) Cmd(cmd string) ([]Response, error) {
	buf := make([]byte, 512)
	var res []byte
	var r []Response

	_, err := c.conn.Write([]byte(cmd + "\r\n"))
	if err != nil {
		return nil, errors.New("Write: " + err.Error())
	}

	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			return nil, errors.New("Read: " + err.Error())
		}

		res = append(res, buf[:n]...)
		if bytes.Contains(res, []byte("\n000")) {
			break
		}
	}

	for _, s := range bytes.Split(res, []byte("\n")) {
		if len(s) == 0 {
			continue
		}

		code, err := strconv.Atoi(string(s[:3]))
		if err != nil {
			return r, errors.New("response code parsing error: " + err.Error())
		}

		if code != 0 {
			r = append(r, Response{code, string(s[4:])})
		}
	}

	return r, nil
}

func (c *Client) Query(dbid, flags, numres int, filename string) ([]QueryResult, error) {
	var results []QueryResult

	responses, err := c.Cmd(fmt.Sprintf("query %d %d %d %s", dbid, flags, numres, filename))
	if err != nil {
		return nil, err
	}

	for _, r := range responses {
		if r.Code != 200 {
			continue
		}

		result := QueryResult{}
		args := strings.Split(r.Content, " ")

		result.ImgID, err = strconv.Atoi(args[0])
		if err != nil {
			return nil, err
		}

		result.Score, err = strconv.ParseFloat(args[1], 64)
		if err != nil {
			return nil, err
		}

		result.Width, err = strconv.Atoi(args[2])
		if err != nil {
			return nil, err
		}

		result.Height, err = strconv.Atoi(args[3])
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
