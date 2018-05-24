package game

import (
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"github.com/pkg/errors"
)

func readMessage(conn net.Connection) (*downstream.Message, error) {
	data, err := conn.Read()
	if err != nil {
		return nil, errors.Wrap(err, "error reading from server")
	}
	message := downstream.GetRootAsMessage(data, 0)
	return message, nil
}
