package grpc

import (
	"fmt"
	"sync"

	api "github.com/Alp4ka/classifier-api"
	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
)

type handler struct {
	inputHandler  *processInputHandler
	outputHandler *processOutputHandler
	stopCh        chan struct{}

	errOnce sync.Once
	errMu   sync.Mutex
	err     error
}

func newHandler(inputChan chan<- *api.ProcessRequest, outputChan <-chan *api.ProcessResponse) *handler {
	return &handler{
		inputHandler:  newProcessInputHandler(inputChan),
		outputHandler: newProcessOutputHandler(outputChan),
		stopCh:        make(chan struct{}, 1),
	}
}

func (h *handler) GetInputHandler() core.ProcessInputHandler {
	return h.inputHandler
}

func (h *handler) GetOutputHandler() core.ProcessOutputHandler {
	return h.outputHandler
}

func (h *handler) Close() {
	h.errOnce.Do(func() {
		close(h.stopCh)
	})
}

func (h *handler) CloseWithError(err error) {
	h.errOnce.Do(func() {
		h.errMu.Lock()
		h.err = err
		h.errMu.Unlock()
	})
}

func (h *handler) Error() error {
	h.errMu.Lock()
	defer h.errMu.Unlock()

	return h.err
}

func (h *handler) Done() <-chan struct{} {
	return h.stopCh
}

type processInputHandler struct {
	inputChan chan<- *api.ProcessRequest
}

func newProcessInputHandler(ch chan<- *api.ProcessRequest) *processInputHandler {
	return &processInputHandler{inputChan: ch}
}

func (h *processInputHandler) Handle(input *core.ProcessInput) error {
	parsed := &api.ProcessRequest{
		RequestId:   input.RequestID.String(),
		RequestData: input.UserInput,
	}

	select {
	case h.inputChan <- parsed:
	default:
		return fmt.Errorf("failed to write to input channel")
	}

	return nil
}

type processOutputHandler struct {
	outputChan <-chan *api.ProcessResponse
}

func newProcessOutputHandler(ch <-chan *api.ProcessResponse) *processOutputHandler {
	return &processOutputHandler{outputChan: ch}
}

func (h *processOutputHandler) HasMessage() bool {
	return len(h.outputChan) != 0
}

func (h *processOutputHandler) Await() (*core.ProcessOutput, error) {
	resp, open := <-h.outputChan
	if !open {
		return nil, fmt.Errorf("output channel closed")
	}

	action, err := protoActionToAction(resp.Action)
	if err != nil {
		return nil, err
	}

	return &core.ProcessOutput{
		Action:       action,
		UserResponse: resp.ResponseData,
	}, nil
}
