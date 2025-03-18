package parser

import (
	"strconv"
	"strings"
)

type Version interface{ version() }

type VersionMachine struct{}

func (v VersionMachine) version() {}

type VersionInterpreter struct {
	Major uint16
	Minor uint16
	Patch uint16
}

func (v VersionInterpreter) version() {}

func (v VersionInterpreter) GtEq(other VersionInterpreter) bool {
	if v.Major > other.Major {
		return true
	}
	if v.Major < other.Major {
		return false
	}
	if v.Minor < other.Minor {
		return false
	}
	return v.Patch >= other.Patch
}

func parseSemanticVersion(src string) (bool, int, int, int) {
	parts := strings.Split(src, ".")
	if len(parts) < 3 {
		return false, 0, 0, 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, 0, 0, 0
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, 0, 0, 0
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, 0, 0, 0
	}
	return true, major, minor, patch
}

func (p Program) GetVersion() Version {
	for _, comment := range p.Comments {
		comment := strings.TrimLeft(comment.Content, " ")
		comment = strings.TrimRight(comment, " \n")

		parts := strings.Split(comment, " ")
		if len(parts) < 2 {
			return nil
		}

		if parts[0] != "@version" {
			return nil
		}

		switch parts[1] {
		case "machine":
			return VersionMachine{}
		case "interpreter":
			ok, major, minor, patch := parseSemanticVersion(parts[2])
			if !ok {
				return nil
			}

			return VersionInterpreter{
				Major: uint16(major),
				Minor: uint16(minor),
				Patch: uint16(patch),
			}
		default:
			return nil

		}

	}

	return nil
}
