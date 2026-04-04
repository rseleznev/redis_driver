package polling

import (
	"testing"
)

var poller Epoller

func TestNewPoller(t *testing.T) {
	var err error
	
	poller, err = NewPoller()
	if err != nil {
		t.Error(err)
	}
}

func TestAdd(t *testing.T) {
	
}