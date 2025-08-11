package web

import (
	"strings"
	"testing"
)

func TestContractLogicIsPointer(t *testing.T) {
	pointerErrorMessage := "request object must be a pointer"
	var a = struct {
		ID string `json:"id" validate:"required" label:"id"`
	}{
		ID: "1234",
	}

	if err := _contractLogic(a); err == nil {
		t.Errorf("expected error, got success")
	} else if !strings.Contains(err.Error(), pointerErrorMessage) {
		t.Errorf("expected: %s, got: %s", pointerErrorMessage, err.Error())
	}
}
