// @Title        types
// @Description  main
// @Create       yiyiyi 2025/11/20 10:29

package sip

type headerType uint

const (
	_ headerType = iota

	headerTypeVia
	headerTypeRegisterFrom
	headerTypeRegisterTo
	headerTypeRegisterCSEq
	headerTypeFrom
	headerTypeTo
	headerTypeToWith
	headerTypeCallId
	headerTypeMessageCSEq
	headerTypeUserAgent
	headerTypeMaxForwards
	headerTypeAuth
	headerTypeOnline
	headerTypeContentType
	headerTypeContentLength
	headerTypeExpire
	headerTypeContact
	headerTypeEventPresence
	headerTypeEventCatalog
	headerTypeCallIdWith
	headerTypeContentTypeSDP
	headerTypeContactCurrent
	headerTypeContentTypeMANSRTSP
	headerTypeSubject
)

const (
	SDPMediaDescription_1 = "v/////a/1/8/1"
)
