package psnats

import nats "github.com/nats-io/nats.go"

func getHeaders(msg *nats.Msg) map[string]string {
	header := map[string]string{}
	for k := range msg.Header {
		header[k] = msg.Header.Get(k)
	}
	return header
}
