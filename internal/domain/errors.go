package domain

import "errors"

var (
	ErrActiveHittersFull        = errors.New("roster already has the maximum active hitters")
	ErrActivePitchersFull       = errors.New("roster already has the maximum active pitchers")
	ErrEventOutsideViewWindow   = errors.New("event is outside view effective window")
	ErrPlayerAlreadyActive      = errors.New("player already activated")
	ErrPlayerAlreadyOnRoster    = errors.New("player already on roster")
	ErrRosterFull               = errors.New("roster is already full")
	ErrPlayerNotOnRoster        = errors.New("player is not on the roster")
	ErrUnrecognizedPlayerRole   = errors.New("unrecognized player role")
	ErrUnrecognizedRosterEvent  = errors.New("unrecognized roster event")
	ErrUnrecognizedRosterStatus = errors.New("unrecognized roster status")
	ErrWrongTeamID              = errors.New("team IDs do not match")
)
