package mock

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	mock "github.com/stretchr/testify/mock"
)


type Terminal struct {
	mock.Mock
}

func (term *Terminal) Init(flags *runtime.Flags) {
	_ = term.Called(flags)
}

func (term *Terminal) SetColors(
