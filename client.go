// Package iqdb provides a client for the iqdb protocol.
package iqdb

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

const buflen = 4096

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
		return nil, fmt.Errorf("iqdb connect %s: %v", addr, err)
	}

	return &Client{conn}, nil
}

func (c *Client) Cmd(cmd string) ([]Response, error) {
	err := c.sendCmd(cmd)
	if err != nil {
		return nil, err
	}

	return c.recvCmd(cmd)
}

func (c *Client) CmdData(cmd string, size int64, r io.Reader) ([]Response, error) {
	err := c.sendCmd(fmt.Sprintf("%s :%d", cmd, size))
	if err != nil {
		return nil, err
	}

	err = c.sendData(r)
	if err != nil {
		return nil, err
	}

	return c.recvCmd(cmd)
}

func (c *Client) sendCmd(cmd string) error {
	_, err := c.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return fmt.Errorf("iqdb cmd \"%s\": %v", cmd, err)
	}

	return nil
}

func (c *Client) sendData(r io.Reader) error {
	buf := make([]byte, buflen)

	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("iqdb payload: %v", err)
		}

		if _, err := c.conn.Write(buf[:n]); err != nil {
			return fmt.Errorf("iqdb payload: %v", err)
		}

		if err == io.EOF {
			c.conn.Write([]byte("\r\n"))
			break
		}
	}

	return nil
}

func (c *Client) recvCmd(cmd string) ([]Response, error) {
	var r []Response
	var res []byte
	buf := make([]byte, buflen)

	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("iqdb cmd \"%s\": %v", cmd, err)
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
			return r, fmt.Errorf("iqdb cmd \"%s\": %v", cmd, err)
		}

		if code != 0 {
			r = append(r, Response{code, string(s[4:])})
		}
	}

	return r, nil
}

func (c *Client) Query(dbid, flags, numres int, filename string) ([]QueryResult, error) {
	responses, err := c.Cmd(fmt.Sprintf("query %d %d %d %s", dbid, flags, numres, filename))
	if err != nil {
		return nil, err
	}

	return c.parseQuery(responses)
}

func (c *Client) QueryData(dbid, flags, numres int, size int64, r io.Reader) ([]QueryResult, error) {
	responses, err := c.CmdData(fmt.Sprintf("query %d %d %d", dbid, flags, numres), size, r)
	if err != nil {
		return nil, err
	}

	return c.parseQuery(responses)
}

func (c *Client) parseQuery(responses []Response) (results []QueryResult, err error) {
	for _, res := range responses {
		if res.Code != ResQuery {
			continue
		}

		r := QueryResult{}
		if _, err = fmt.Sscanf(res.Content, "%x %f %d %d", &r.ImgID, &r.Score, &r.Width, &r.Height); err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	return results, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
