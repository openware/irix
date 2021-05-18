package cryptocom

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestWsHeartbeat(t *testing.T) {
	testCases := []struct {
		id        int
		isPrivate bool
	}{
		{int(timestampMs(time.Now())), false},
		{int(timestampMs(time.Now())), true},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if c.isPrivate {
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(nil)
		} else {
			public.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(nil)
		}
		cli.respondHeartBeat(c.isPrivate, c.id)
		var pr Request
		if c.isPrivate {
			private.AssertExpectations(t)
			private.AssertNumberOfCalls(t, "WriteMessage", 1)
			public.AssertNumberOfCalls(t, "WriteMessage", 0)
			req := private.Calls[0].Arguments[1].([]byte)
			_ = json.Unmarshal(req, &pr)
		} else {
			public.AssertExpectations(t)
			public.AssertNumberOfCalls(t, "WriteMessage", 1)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			req := public.Calls[0].Arguments[1].([]byte)
			_ = json.Unmarshal(req, &pr)
		}
		assert.Equal(t, publicRespondHeartbeat, pr.Method)
		assert.Equal(t, c.id, pr.Id)
		assert.Empty(t, pr.Signature)
		assert.Empty(t, pr.ApiKey)
	}
}
