package router

import (
    "errors"
)
var ErrConnectingToGeth = errors.New("Error connecting to geth!")
var ErrReadingGethResponse = errors.New("Error reading response from Geth!")
var ErrParsingJSON = errors.New("Error while parsing JSON!")
var ErrEncodingJSON = errors.New("Error while encoding JSON!")
var ErrParsingInt = errors.New("Error while parsing Int!")
var ErrNullResult = errors.New("Error! Geth returned NULL result!")